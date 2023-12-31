
import json
def flatten_json(json_obj, prefix=''):
    flat_dict = {}
    for key, value in json_obj.items():
        if isinstance(value, dict):
            flat_dict.update(flatten_json(value, f"{prefix}{key}_"))
        elif isinstance(value, list) and len(value) > 0 and isinstance(value[0], dict):
            for i, item in enumerate(value):
                flat_dict.update(flatten_json(item, f"{prefix}{key}_{i}_"))
        else:
            flat_dict[f"{prefix}{key}"] = value
    return flat_dict

# Input JSON
input_json = '''
{
   "filter":{
      "sitereference":[
         {
            "value":"shshs"
         }
      ],
      "transactionreference":[
         {
            "value":"3-64-25205"
         }
      ]
   }
}
'''

# Load JSON
data = json.loads(input_json)

# Flatten JSON without changing data types
flattened_data = flatten_json(data)

# Print flattened JSON
print(json.dumps(flattened_data, indent=2))