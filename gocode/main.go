package main

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

type JSONToGoConverter struct {
	Data          interface{}
	Go            string
	Tabs          int
	Seen          map[string][]string
	Stack         []string
	Accumulator   string
	InnerTabs     int
	Parent        string
	TypeName      string
	Flatten       bool
	Example       bool
	AllOmitEmpty  bool
	BSON          bool
	BSONOmitEmpty bool
}

func (c *JSONToGoConverter) NewJSONToGoConverter(jsonStr string, typeName string, flatten bool, example bool, allOmitEmpty bool, bson bool, bsonOmitEmpty bool) *JSONToGoConverter {
	var data interface{}
	err := json.Unmarshal([]byte(jsonStr), &data)
	if err != nil {
		panic(err)
	}

	return &JSONToGoConverter{
		Data:          data,
		Go:            "",
		Tabs:          0,
		Seen:          make(map[string][]string),
		Stack:         make([]string, 0),
		Accumulator:   "",
		InnerTabs:     0,
		Parent:        "",
		TypeName:      c.Format(typeName),
		Flatten:       flatten,
		Example:       example,
		AllOmitEmpty:  allOmitEmpty,
		BSON:          bson,
		BSONOmitEmpty: bsonOmitEmpty,
	}
}

func (c *JSONToGoConverter) Convert() string {
	c.Append(fmt.Sprintf("type %s ", c.TypeName))
	c.ParseScope(c.Data, 0)
	if c.Flatten {
		return c.Go + c.Accumulator
	} else {
		return c.Go
	}
}

func (c *JSONToGoConverter) ParseScope(scope interface{}, depth int) {
	switch v := scope.(type) {
	case map[string]interface{}:
		if c.Flatten {
			if depth >= 2 {
				c.Appender(c.Parent)
			} else {
				c.Append(c.Parent)
			}
			c.ParseStruct(depth+1, c.InnerTabs, v, nil)
		}
	case []interface{}:
		sliceType := ""
		for _, item := range v {
			thisType := c.GoType(item)
			if sliceType == "" {
				sliceType = thisType
			} else if sliceType != thisType {
				sliceType = c.MostSpecificPossibleGoType(thisType, sliceType)
				if sliceType == "any" {
					break
				}
			}
		}
		b := []string{"struct", "slice"}

		if c.Flatten && slices.Contains(b, sliceType) {
			sliceStr := fmt.Sprintf("[]%s", c.Parent)
			if c.InnerTabs >= 2 {
				c.Appender(sliceStr)
			} else {
				c.Append(sliceStr)
			}
		} else {
			sliceStr := "[]"
			if c.Flatten && c.InnerTabs >= 2 {
				c.Appender(sliceStr)
			} else {
				c.Append(sliceStr)
			}

			if sliceType == "struct" {
				allFields := make(map[string]map[string]interface{})
				for _, item := range v {
					for key, value := range item.(map[string]interface{}) {
						if _, ok := allFields[key]; !ok {
							allFields[key] = map[string]interface{}{"value": value, "count": 0}
						} else {
							existingValue := allFields[key]["value"]
							currentValue := value
							if c.CompareObjects(existingValue, currentValue) {
								comparisonResult := c.CompareObjectKeys(
									extractKeys(reflect.ValueOf(allFields).MapKeys()),
									extractKeys(reflect.ValueOf(allFields).MapKeys()),
								)
								if !comparisonResult {
									key = fmt.Sprintf("%s_%s", key, c.UUIDv4())
									allFields[key] = map[string]interface{}{"value": currentValue, "count": 0}
								}
							}
							// Convert the count value to int before incrementing
							count := allFields[key]["count"].(int)
							count += 1
							allFields[key]["count"] = count
						}
					}
				}
				structKeys := extractKeys(reflect.ValueOf(allFields).MapKeys())
				structMap := make(map[string]interface{})
				omitemptyMap := make(map[string]bool)

				for _, key := range structKeys {
					elem := allFields[key]
					structMap[key] = elem["value"]
					omitemptyMap[key] = elem["count"] != len(v)
				}
				c.ParseStruct(depth+1, c.InnerTabs, structMap, omitemptyMap)
			} else if sliceType == "slice" {
				c.ParseScope(v[0], depth)
			} else {
				if c.Flatten && depth >= 2 {
					c.Appender(sliceType)
				} else {
					c.Append(sliceType)
				}
			}
		}
	default:
		if c.Flatten && depth >= 2 {
			c.Appender(c.GoType(v))
		} else {
			c.Append(c.GoType(v))
		}
	}
}

func (c *JSONToGoConverter) ParseStruct(depth int, innerTabs int, scope map[string]interface{}, omitempty map[string]bool) {
	if c.Flatten {
		c.Stack = append(c.Stack, "\n")
	}
	seenTypeNames := make([]string, 0)

	if c.Flatten && depth >= 2 {
		parentType := fmt.Sprintf("type %s ", c.Parent)
		scopeKeys := c.formatScopeKeys(extractKeys(reflect.ValueOf(scope).MapKeys()))
		if _, ok := c.Seen[c.Parent]; ok && c.CompareObjectKeys(scopeKeys, c.Seen[c.Parent]) {
			c.Stack = c.Stack[:len(c.Stack)-1]
			return
		}
		c.Seen[c.Parent] = scopeKeys

		c.Appender(fmt.Sprintf("%s\t struct {\n", parentType))
		c.InnerTabs += 1
		keys := extractKeys(reflect.ValueOf(scope).MapKeys())
		fmt.Println(keys)
		for _, key := range keys {
			keyName := c.GetOriginalName(key)
			c.Indenter(c.InnerTabs)
			typeName := c.UniqueTypeName(c.Format(keyName), seenTypeNames)
			seenTypeNames = append(seenTypeNames, typeName)
			c.Appender(fmt.Sprintf("%s ", typeName))
			c.Parent = typeName
			c.ParseScope(scope[key], depth)
			c.Appender(fmt.Sprintf(" `json:\"%s", keyName))

			if c.AllOmitEmpty || (c.AllOmitEmpty && omitempty[key]) {
				c.Appender(",omitempty")
			}
			if c.BSON {
				c.Appender("\" bson:\"" + keyName)
			}
			if c.BSONOmitEmpty {
				c.Appender(",omitempty")
			}
			c.Appender("\"`\n")
		}
		c.Indenter(c.InnerTabs - 1)
		c.Appender("}")
	} else {
		c.Append("struct {\n")
		c.Tabs += 1
		keys := extractKeys(reflect.ValueOf(scope).MapKeys())
		for _, key := range keys {
			keyName := c.GetOriginalName(key)
			c.Indent(c.Tabs)
			typeName := c.UniqueTypeName(c.Format(keyName), seenTypeNames)
			seenTypeNames = append(seenTypeNames, typeName)
			c.Append(fmt.Sprintf("%s ", typeName))
			c.Parent = typeName
			c.ParseScope(scope[key], depth)
			c.Append(fmt.Sprintf(" `json:\"%s", keyName))

			if c.AllOmitEmpty || (c.AllOmitEmpty && omitempty[key]) {
				c.Append(",omitempty")
			}
			if c.BSON {
				c.Append("\" bson:\"" + keyName)
			}
			if c.BSONOmitEmpty {
				c.Append(",omitempty")
			}
			c.Append("\"`\n")
		}
		c.Indent(c.Tabs - 1)
		c.Append("}")
	}

	if c.Flatten {
		c.Accumulator += c.Stack[len(c.Stack)-1]
		// c.Stack = c.Stack[:len(c.Stack)-1]
	}
}

func (c *JSONToGoConverter) Indent(tabs int) {
	c.Appender(strings.Repeat("\t", tabs))
}

func (c *JSONToGoConverter) Append(str string) {
	c.Go += str
}

func (c *JSONToGoConverter) Indenter(tabs int) {
	c.Stack[len(c.Stack)-1] += strings.Repeat("\t", tabs)
}

func (c *JSONToGoConverter) Appender(str string) {
	c.Stack[len(c.Stack)-1] += str
}

func (c *JSONToGoConverter) UniqueTypeName(name string, seen []string) string {
	if !contains(seen, name) {
		return name
	}

	i := 0
	for {
		newName := name + strconv.Itoa(i)
		if !contains(seen, newName) {
			return newName
		}
		i++
	}
}

func (c *JSONToGoConverter) Format(str string) string {
	str = c.FormatNumber(str)

	sanitized := c.ToProperCase(str)
	re := regexp.MustCompile("[^a-zA-Z0-9]")
	sanitized = re.ReplaceAllString(sanitized, "")

	if sanitized == "" {
		return "NAMING_FAILED"
	}

	// After sanitizing, the remaining characters can start with a number.
	// Run the sanitized string again through formatNumber to ensure the identifier is Num[0-9] or Zero_... instead of 1.
	return c.FormatNumber(sanitized)
}

func (c *JSONToGoConverter) FormatNumber(str string) string {
	if str == "" {
		return ""
	} else if matched, _ := regexp.MatchString(`^\d+$`, str); matched {
		str = "Num" + str
	} else if strings.IndexAny(string(str[0]), "0123456789") != -1 {
		numbers := map[string]string{
			"0": "Zero_", "1": "One_", "2": "Two_", "3": "Three_",
			"4": "Four_", "5": "Five_", "6": "Six_", "7": "Seven_",
			"8": "Eight_", "9": "Nine_",
		}
		str = numbers[string(str[0])] + str[1:]
	}

	return str
}

func (c *JSONToGoConverter) GoType(val interface{}) string {
	switch v := val.(type) {
	case bool:
		return "bool"
	case string:
		if c.IsDatetimeString(v) {
			return "time.Time"
		}
		return "string"
	case int:
		if -2147483648 < v && v < 2147483647 {
			return "int"
		}
		return "int64"
	case float64:
		if float64(int(v)) == v {
			return "int"
		}
		return "float64"
	case []interface{}:
		return "slice"
	case map[string]interface{}:
		return "struct"
	default:
		return "any"
	}
}

func (c *JSONToGoConverter) MostSpecificPossibleGoType(typ1 string, typ2 string) string {
	if strings.HasPrefix(typ1, "float") && strings.HasPrefix(typ2, "int") {
		return typ1
	} else if strings.HasPrefix(typ1, "int") && strings.HasPrefix(typ2, "float") {
		return typ2
	} else {
		return "any"
	}
}

func (c *JSONToGoConverter) ToProperCase(str string) string {
	// Ensure that the SCREAMING_SNAKE_CASE is converted to snake_case
	if match, _ := regexp.MatchString("^[_A-Z0-9]+$", str); match {
		str = strings.ToLower(str)
	}

	// List of common initialisms
	commonInitialisms := map[string]bool{
		"ACL": true, "API": true, "ASCII": true, "CPU": true, "CSS": true, "DNS": true,
		"EOF": true, "GUID": true, "HTML": true, "HTTP": true, "HTTPS": true, "ID": true,
		"IP": true, "JSON": true, "LHS": true, "QPS": true, "RAM": true, "RHS": true,
		"RPC": true, "SLA": true, "SMTP": true, "SQL": true, "SSH": true, "TCP": true,
		"TLS": true, "TTL": true, "UDP": true, "UI": true, "UID": true, "UUID": true,
		"URI": true, "URL": true, "UTF8": true, "VM": true, "XML": true, "XMPP": true,
		"XSRF": true, "XSS": true,
	}

	// Convert the string to Proper Case
	re := regexp.MustCompile(`(^|[^a-zA-Z])([a-z]+)`)
	str = re.ReplaceAllStringFunc(str, func(match string) string {
		parts := re.FindStringSubmatch(match)
		sep, frag := parts[1], parts[2]

		if commonInitialisms[strings.ToUpper(frag)] {
			return sep + strings.ToUpper(frag)
		} else {
			return sep + strings.ToUpper(frag[0:1]) + strings.ToLower(frag[1:])
		}
	})

	re = regexp.MustCompile(`([A-Z])([a-z]+)`)
	str = re.ReplaceAllStringFunc(str, func(match string) string {
		parts := re.FindStringSubmatch(match)
		sep, frag := parts[1], parts[2]

		if commonInitialisms[sep+strings.ToUpper(frag)] {
			return (sep + frag)[0:]
		} else {
			return sep + frag
		}
	})

	return str
}

func (c *JSONToGoConverter) UUIDv4() string {
	uuid := make([]byte, 16)
	_, err := rand.Read(uuid)
	if err != nil {
		panic(err)
	}

	// Set version (4) and variant bits (2)
	uuid[6] = (uuid[6] & 0x0F) | 0x40
	uuid[8] = (uuid[8] & 0x3F) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}

func (c *JSONToGoConverter) GetOriginalName(unique string) string {
	reLiteralUUID := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	uuidLength := 36

	if len(unique) >= uuidLength {
		tail := unique[len(unique)-uuidLength:]
		if reLiteralUUID.MatchString(tail) {
			return unique[:len(unique)-(uuidLength+1)]
		}
	}
	return unique
}

func (c *JSONToGoConverter) CompareObjects(objectA, objectB interface{}) bool {
	typeObject := reflect.TypeOf(map[string]interface{}{})

	return reflect.TypeOf(objectA) == typeObject &&
		reflect.TypeOf(objectB) == typeObject
}

func (c *JSONToGoConverter) CompareObjectKeys(itemAKeys []string, itemBKeys []string) bool {
	lengthA, lengthB := len(itemAKeys), len(itemBKeys)

	if lengthA == 0 && lengthB == 0 {
		return true
	}

	if lengthA != lengthB {
		return false
	}

	for _, item := range itemAKeys {
		if !contains(itemBKeys, item) {
			return false
		}
	}

	return true
}

func (c *JSONToGoConverter) FormatScopeKeys(keys []string) []string {
	for i := range keys {
		keys[i] = c.Format(keys[i])
	}
	return keys
}

func extractKeys(keys []reflect.Value) []string {
	result := make([]string, len(keys))
	for i, key := range keys {
		result[i] = key.Interface().(string)
	}
	return result
}

func (c *JSONToGoConverter) IsDatetimeString(str string) bool {
	_, err := time.Parse(time.RFC3339, str)
	return err == nil
}

func (c *JSONToGoConverter) formatScopeKeys(keys []string) []string {
	for i := range keys {
		keys[i] = c.Format(keys[i])
	}
	return keys
}

func isDigit(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

func main() {
	jsonInput := ``
	typeName := "EventData1"
	b := JSONToGoConverter{}

	converter := b.NewJSONToGoConverter(jsonInput, typeName, true, false, true, false, false)
	result := converter.Convert()
	goCode := fmt.Sprintf("package main\n\n%s", result)

	// Write the Go code to a file
	filePath := "generated_struct.go"
	err := ioutil.WriteFile(filePath, []byte(goCode), 0644)
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return
	}

	fmt.Printf("Go code written to %s\n", filePath)
}
