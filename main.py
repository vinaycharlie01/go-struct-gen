import json
import re
import random
from datetime import datetime
import json
import logging

class JSONToGoConverter:
    def __init__(self, json_str, typename="AutoGenerated", flatten=True, example=False, all_omitempty=False,bson=False,bson_omitempty=False):
        self.data = json.loads(json_str.replace(r'(:\s*\[?\s*-?\d*)\.0', r'\1.1'))
        self.scope = self.data
        self.go = ""
        self.tabs = 0
        self.seen = {}
        self.stack = []
        self.accumulator = ""
        self.inner_tabs = 0
        self.parent = ""
        self.typename = self.format(typename or "AutoGenerated")
        self.flatten = flatten
        self.example = example
        self.all_omitempty = all_omitempty
        self.bson=bson
        self.bson_omitempty=bson_omitempty
        logging.info(json.dumps({
            "data":self.data,
            "typename": self.format(typename or "AutoGenerated")
        }))

    def convert(self):
        self.append(f"type {self.typename} ")
        self.parse_scope(self.scope)
        return self.go + self.accumulator if self.flatten else self.go
        # return self.go

    def parse_scope(self, scope, depth=0):
        if isinstance(scope, dict):
            if self.flatten:
                if depth>=2:
                    self.appender(self.parent)
                else:
                    self.append(self.parent)
            self.parse_struct(depth + 1, self.inner_tabs, scope)
        elif isinstance(scope, list):
            slice_type = None
            for item in scope:
                this_type = self.go_type(item)
                if not slice_type:
                    slice_type = this_type
                elif slice_type != this_type:
                    slice_type = self.most_specific_possible_go_type(this_type, slice_type)
                    if slice_type == "any":
                        break
            # slice_str = f"[]{self.parent}" if self.flatten and slice_type in ["struct", "slice"] else "[]"
            # slice_str = f"[]"
            if self.flatten and slice_type in ["struct","slice"]:
                slice_str = f"[]{self.parent}"
            else:
                slice_str = f"[]" 
            if self.flatten and depth >= 2:
                self.appender(slice_str)
            else:
                self.append(slice_str)

            if slice_type == "struct":
                all_fields = {}
                for item in scope:
                    for key, value in item.items():
                        if key not in all_fields:
                            all_fields[key] = {"value": value, "count": 0}
                        else:
                            existing_value = all_fields[key]["value"]
                            current_value = value
                            if self.compare_objects(existing_value, current_value):
                                comparison_result = self.compare_object_keys(
                                    list(current_value.keys()),
                                    list(existing_value.keys())
                                )
                                if not comparison_result:
                                    key = f"{key}_{self.uuidv4()}"
                                    all_fields[key] = {"value": current_value, "count": 0}
                            all_fields[key]["count"] += 1

                struct_keys = list(all_fields.keys())
                struct = {}
                omitempty = {}
                for key in struct_keys:
                    elem = all_fields[key]
                    struct[key] = elem["value"]
                    omitempty[key] = elem["count"] != len(scope)

                self.parse_struct(depth + 1, self.inner_tabs, struct, omitempty)
            elif slice_type == "slice":
                self.parse_scope(scope[0], depth)
            else:
                if self.flatten and depth >= 2:
                    self.appender(slice_type or "any")
                else:
                    self.append(slice_type or "any")
        else:
            if self.flatten and depth >= 2:
                self.appender(self.go_type(scope))
            else:
                self.append(self.go_type(scope))

    def parse_struct(self, depth, inner_tabs, scope, omitempty=None):
        if self.flatten:
            self.stack.append("\n" if depth >= 2 else "")
        seen_type_names = []
        if self.flatten and depth >= 2:
            parent_type = f"type {self.parent}"
            scope_keys = self.format_scope_keys(list(scope.keys()))
            if self.parent in self.seen and self.compare_object_keys(scope_keys, self.seen[self.parent]):
                self.stack.pop()
                return
            self.seen[self.parent] = scope_keys

            self.appender(f"{parent_type}"+" "+"struct {\n")
            self.inner_tabs += 1
            keys = list(scope.keys())
            for key in keys:
                keyname = self.get_original_name(key)
                self.indenter(self.inner_tabs)
                typename = self.unique_type_name(self.format(keyname), seen_type_names)
                seen_type_names.append(typename)
                self.appender(f"{typename}"+" ")
                self.parent = typename
                self.parse_scope(scope[key], depth)
                self.appender(' `json:"'+keyname)

                if self.all_omitempty or (self.all_omitempty and omitempty[key] == True):
                    self.appender(',omitempty')
                if self.bson:
                    self.appender('" bson:"' + keyname)
                if self.bson_omitempty:
                    self.appender(',omitempty')
                self.appender('"`\n')
            self.indenter(self.inner_tabs - 1)
            self.appender("}")
        else:
            self.append("struct {\n")
            self.tabs += 1
            keys = list(scope.keys())
            for key in keys:
                keyname = self.get_original_name(key)
                self.indent(self.tabs)
                typename = self.unique_type_name(self.format(keyname), seen_type_names)
                seen_type_names.append(typename)
                self.append(typename+" ")
                self.parent = typename
                self.parse_scope(scope[key], depth)
                self.append(' `json:"'+keyname)
                if self.all_omitempty or (self.all_omitempty and omitempty[key] == True):
                    self.append(',omitempty')
                if self.bson:
                    self.append('" bson:"' + keyname)
                if self.bson_omitempty:
                    self.append(',omitempty')
                self.append('"`\n')

            self.indent(self.tabs - 1)
            self.append("}")

        if self.flatten:
            self.accumulator += self.stack.pop()

    def indent(self, tabs):
        self.append('\t' * tabs)

    def append(self, string):
        self.go += string

    def indenter(self, tabs):
        self.stack[-1] += '\t' * tabs

    def appender(self, string):
        self.stack[-1] += string

    def unique_type_name(self, name, seen):
        if name not in seen:
            return name

        i = 0
        while True:
            new_name = f"{name}{i}"
            if new_name not in seen:
                return new_name
            i += 1

    def format(self, string):
        string = self.format_number(string)

        sanitized = self.to_proper_case(string).replace(r'[^a-z0-9]', "")
        if not sanitized:
            return "NAMING_FAILED"

        return self.format_number(sanitized)

    def format_number(self, string):
        if not string:
            return ""
        elif string.isdigit():
            string = f"Num{string}"
        elif string[0].isdigit():
            numbers = {'0': "Zero_", '1': "One_", '2': "Two_", '3': "Three_",
                       '4': "Four_", '5': "Five_", '6': "Six_", '7': "Seven_",
                       '8': "Eight_", '9': "Nine_"}
            string = numbers[string[0]] + string[1:]

        return string

    def go_type(self, val):
        if isinstance(val, bool):
            return "bool"
        if isinstance(val, str):
            if re.match(r'^\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d(\.\d+)?(\+\d\d:\d\d|Z)$', val):
                return "time.Time"
            else:
                return "string"
        if isinstance(val, (int, float)):
            if isinstance(val, int) and -2147483648 < val < 2147483647:
                return "int"
            else:
                return "int64" if isinstance(val, int) else "float64"
        if isinstance(val, list):
            # b = {}
            # for value in val:
            #     if isinstance(value, list):
            #         for item in value:
            #             b[str(type(item).__name__)] = type(item).__name__
            # if len(b)==1:
            #     return "[]"+str(type(val[0]).__name__)
            # else:
            #     return "[]interface{}"
            return "slice"
        elif isinstance(val, dict):
            return "struct"
        else:
            return "any"

    def most_specific_possible_go_type(self, typ1, typ2):
        if typ1[:5] == "float" and typ2[:3] == "int":
            return typ1
        elif typ1[:3] == "int" and typ2[:5] == "float":
            return typ2
        else:
            return "any"

    def to_proper_case(self, s):
        # Ensure that the SCREAMING_SNAKE_CASE is converted to snake_case
        if re.match("^[_A-Z0-9]+$", s):
            s = s.lower()

        # List of common initialisms
        common_initialisms = {
            "ACL", "API", "ASCII", "CPU", "CSS", "DNS",
            "EOF", "GUID", "HTML", "HTTP", "HTTPS", "ID",
            "IP", "JSON", "LHS", "QPS", "RAM", "RHS",
            "RPC", "SLA", "SMTP", "SQL", "SSH", "TCP",
            "TLS", "TTL", "UDP", "UI", "UID", "UUID",
            "URI", "URL", "UTF8", "VM", "XML", "XMPP",
            "XSRF", "XSS",
        }

        # Convert the string to Proper Case
        s = re.sub(r'(^|[^a-zA-Z])([a-z]+)', lambda match: match.group(1) + match.group(2).upper() if match.group(2).upper() in common_initialisms else match.group(1) + match.group(2).capitalize(), s)

        s = re.sub(r'([A-Z])([a-z]+)', lambda match: match.group(1) + match.group(2) if match.group(1) + match.group(2).upper() in common_initialisms else match.group(1) + match.group(2), s)
        s = re.sub(r'[^a-zA-Z0-9_]', '', s)
        s = s.replace('_', '')

        return s
    def uuidv4(self):
        return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(r'[xy]', lambda c: (str(hex(int(c, 16) & 0xf))[2:] if c == 'x' else str(hex(int(c, 16) & 0x3 | 0x8))[2:]))

    def get_original_name(self, unique):
        re_literal_uuid = r'^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$'
        uuid_length = 36

        if len(unique) >= uuid_length:
            tail = unique[-uuid_length:]
            if re.match(re_literal_uuid, tail):
                return unique[:-1 * (uuid_length + 1)]
        return unique

    def compare_objects(self, object_a, object_b):
        return isinstance(object_a, dict) and isinstance(object_b, dict)

    def compare_object_keys(self, item_a_keys, item_b_keys):
        length_a = len(item_a_keys)
        length_b = len(item_b_keys)

        if length_a == 0 and length_b == 0:
            return True

        if length_a != length_b:
            return False

        for item in item_a_keys:
            if item not in item_b_keys:
                return False

        return True

    def format_scope_keys(self, keys):
        for i in range(len(keys)):
            keys[i] = self.format(keys[i])
        return keys

    def is_datetime_string(self, string):
        try:
            datetime.fromisoformat(string)
            return True
        except ValueError:
            return False


# # Example Usage        
json_input='{ "id": 1296269, "owner": { "login": "octocat", "id": 1, "avatar_url": "https://github.com/images/error/octocat_happy.gif", "gravatar_id": "somehexcode", "url": "https://api.github.com/users/octocat", "html_url": "https://github.com/octocat", "followers_url": "https://api.github.com/users/octocat/followers", "following_url": "https://api.github.com/users/octocat/following{/other_user}", "gists_url": "https://api.github.com/users/octocat/gists{/gist_id}", "starred_url": "https://api.github.com/users/octocat/starred{/owner}{/repo}", "subscriptions_url": "https://api.github.com/users/octocat/subscriptions", "organizations_url": "https://api.github.com/users/octocat/orgs", "repos_url": "https://api.github.com/users/octocat/repos", "events_url": "https://api.github.com/users/octocat/events{/privacy}", "received_events_url": "https://api.github.com/users/octocat/received_events", "type": "User", "site_admin": false, "listdata":[1,2,3,4], "hello":["heher","rnrnrn",1]}}'
typename = "EventData1"


logging.basicConfig(filename='logfile.log', level=logging.INFO, format='%(message)s')


converter = JSONToGoConverter(json_input, typename, flatten=True, example=False, all_omitempty=True,bson=False,bson_omitempty=False)
result = converter.convert()
go_code = f'package main\n\n{result}'

# Write the Go code to a file
file_path = 'generated_struct.go'
with open(file_path, 'w') as file:
    file.write(go_code)

print(f'Go code written to {file_path}')



# # Example Usage        
json_input='{ "type": "Feature", "geometry": { "type": "Point", "coordinates": [ 12.51, 54.67, 0 ] }, "properties": { "meta": { "updated_at": "2022-07-23T18:38:24Z", "units": { "air_pressure_at_sea_level": "hPa", "air_temperature": "celsius", "cloud_area_fraction": "%", "precipitation_amount": "mm", "relative_humidity": "%", "wind_from_direction": "degrees", "wind_speed": "m/s" } }, "timeseries": [ { "time": "2022-07-23T19:00:00Z", "data": { "instant": { "details": { "air_pressure_at_sea_level": 1016.7, "air_temperature": 17.8, "cloud_area_fraction": 99.5, "relative_humidity": 71.6, "wind_from_direction": 289.4, "wind_speed": 6.2 } }, "next_12_hours": { "summary": { "symbol_code": "partlycloudy_day" } }, "next_1_hours": { "summary": { "symbol_code": "cloudy" }, "details": { "precipitation_amount": 0.0 } }, "next_6_hours": { "summary": { "symbol_code": "partlycloudy_night" }, "details": { "precipitation_amount": 0.0 } } } } ] } }'
typename = "EventData2"


logging.basicConfig(filename='logfile.log', level=logging.INFO, format='%(message)s')


converter = JSONToGoConverter(json_input, typename, flatten=True, example=False, all_omitempty=True,bson=False,bson_omitempty=False)
result = converter.convert()
go_code = f'package main\n\n{result}'

# Write the Go code to a file
file_path = 'generated_struct2.go'
with open(file_path, 'w') as file:
    file.write(go_code)

print(f'Go code written to {file_path}')




json_input='[ { "input_index": 0, "candidate_index": 0, "delivery_line_1": "1 N Rosedale St", "last_line": "Baltimore MD 21229-3737", "delivery_point_barcode": "212293737013", "components": { "primary_number": "1", "street_predirection": "N", "street_name": "Rosedale", "street_suffix": "St", "city_name": "Baltimore", "state_abbreviation": "MD", "zipcode": "21229", "plus4_code": "3737", "delivery_point": "01", "delivery_point_check_digit": "3" }, "metadata": { "record_type": "S", "zip_type": "Standard", "county_fips": "24510", "county_name": "Baltimore City", "carrier_route": "C047", "congressional_district": "07", "rdi": "Residential", "elot_sequence": "0059", "elot_sort": "A", "latitude": 39.28602, "longitude": -76.6689, "precision": "Zip9", "time_zone": "Eastern", "utc_offset": -5, "dst": true }, "analysis": { "dpv_match_code": "Y", "dpv_footnotes": "AABB", "dpv_cmra": "N", "dpv_vacant": "N", "active": "Y" } }, { "input_index": 0, "candidate_index": 1, "delivery_line_1": "1 S Rosedale St", "last_line": "Baltimore MD 21229-3739", "delivery_point_barcode": "212293739011", "components": { "primary_number": "1", "street_predirection": "S", "street_name": "Rosedale", "street_suffix": "St", "city_name": "Baltimore", "state_abbreviation": "MD", "zipcode": "21229", "plus4_code": "3739", "delivery_point": "01", "delivery_point_check_digit": "1" }, "metadata": { "record_type": "S", "zip_type": "Standard", "county_fips": "24510", "county_name": "Baltimore City", "carrier_route": "C047", "congressional_district": "07", "rdi": "Residential", "elot_sequence": "0064", "elot_sort": "A", "latitude": 39.2858, "longitude": -76.66889, "precision": "Zip9", "time_zone": "Eastern", "utc_offset": -5, "dst": true }, "analysis": { "dpv_match_code": "Y", "dpv_footnotes": "AABB", "dpv_cmra": "N", "dpv_vacant": "N", "active": "Y" } } ]'
typename = "EventData3"


logging.basicConfig(filename='logfile.log', level=logging.INFO, format='%(message)s')


converter = JSONToGoConverter(json_input, typename, flatten=True, example=False, all_omitempty=True,bson=False,bson_omitempty=False)
result = converter.convert()
go_code = f'package main\n\n{result}'

# Write the Go code to a file
file_path = 'generated_struct3.go'
with open(file_path, 'w') as file:
    file.write(go_code)

print(f'Go code written to {file_path}')
























# # Specify the path to your JSON file
# json_file_path = './hello.json'

# # Read the content of the JSON file
# with open(json_file_path, 'r') as json_file:
#     json_input =json_file.read()
#     # print(json_input)

# typename = "EventData"
# # Use the JSON content as input for the converter
# converter = JSONToGoConverter(json_input, typename, flatten=True, example=False, all_omitempty=True, bson=True, bson_omitempty=True)
# result = converter.convert()
# go_code = f'package main\n\n{result}'

# # Write the Go code to a file
# file_path = 'generated_struct.go'
# with open(file_path, 'w') as file:
#     file.write(go_code)

# print(f'Go code written to {file_path}')
# # app = Flask(__name__)

# from flask import Flask, request, jsonify
# @app.route('/generate_go_code', methods=['POST'])
# def generate_go_code():
#     try:
#         json_data = request.get_json()
#         typename = request.args.get('typename', 'AutoGenerated')
#         flatten = request.args.get('flatten', 'true').lower() == 'true'
#         example = request.args.get('example', 'false').lower() == 'true'
#         all_omitempty = request.args.get('all_omitempty', 'true').lower() == 'true'
#         bson = request.args.get('bson', 'false').lower() == 'true'
#         bson_omitempty = request.args.get('bson_omitempty', 'false').lower() == 'true'

#         converter = JSONToGoConverter(json.dumps(json_data), typename,
#                                       flatten=flatten, example=example,
#                                       all_omitempty=all_omitempty, bson=bson,
#                                       bson_omitempty=bson_omitempty)
#         result = converter.convert()
        
#         go_code = f'package main\n\n{result}'
        
#         return  go_code
#     except Exception as e:
#         return jsonify({'success': False, 'error': str(e)})

# import webbrowser

# # def open_html_page(html_file_path):
# #     webbrowser.open(f'file://{html_file_path}', new=2)

# # if __name__ == "__main__":
# #     # Replace 'index.html' with the actual path to your HTML file
# #     html_file_path = './index.html'
# #     open_html_page(html_file_path)


# if __name__ == '__main__':
#     app.run(debug=True)



# # http://127.0.0.1:5000/generate_go_code?typename=MyType&flatten=false&example=true&all_omitempty=false&bson=true&bson_omitempty=true"
