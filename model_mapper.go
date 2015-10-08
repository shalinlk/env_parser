/**
env_parser performs parsing of the environment variable
and sets the passed struct with value read from the
environment.
Value for each of the struct variable is set by searching
the corresponding name tag in the environment.

Author : shalin LK <shalinkktl@gmail.com>
 */

package env_parser

import (
	"errors"
	"github.com/shalinkktl/env_parser/models"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	APP_TAG   = "env"
	OPTIONAL  = "optional"
	MANDATORY = "mandatory"
)

//EnvToStruct is the core object of the environment
//parser on which mapping function is defined.
type EnvToStruct struct {
	appName   string
	separator string
}

//NewEnvParser returns an object of EnvToStruct
func NewEnvParser() EnvToStruct {
	e := EnvToStruct{}
	return e
}

//Name allows to set name of the application.
//If name is set, parser will append the name of
//the application as prefix with each name tag of
//the variable in the struct.
func (e *EnvToStruct) Name(name string) {
	e.appName = name
}

//Separator appends the supplied prefix to be used
//as a separator for the app name and the tag name of
//the variable.
//Eg app name being 'demo' and separator being '_'
//then the phrase 'demo_' will be prefixed with
//each name tag of the variable in the struct.
func (e *EnvToStruct) Separator(sep string) {
	e.separator = sep
}

//Map performs the mapping of the environment value
//to the struct.
//source is the structure to which environment values
//are to be mapped.
func (e EnvToStruct) Map(source interface{}) error {

	if reflect.ValueOf(source).IsNil() {
		return errors.New("Nil object received for mapping")
	}

	rType := reflect.TypeOf(source)
	if rType.Kind() != reflect.Ptr {
		return errors.New("Expects pointer to target object")
	}
	rType = rType.Elem()
	rFieldNum := rType.NumField()
	metaHolder := make(map[int]fieldMeta, rFieldNum)

	for i := 0; i < rFieldNum; i++ {
		rTFi := rType.Field(i)
		thisMeta := fieldMeta{}
		thisMeta.position = i
		rFTag := rTFi.Tag.Get(APP_TAG)
		tagInfo, tagEr := tagParser(rFTag)
		if tagEr != nil {
			return *tagEr
		}
		thisMeta.envName = tagInfo.name
		thisMeta.mandatory = tagInfo.mandatory
		thisMeta.defVal = tagInfo.deftVal
		metaHolder[i] = thisMeta
	}
	metaHolder = e.envToHolder(e.appName + e.separator, metaHolder)

	rValue := reflect.ValueOf(source)
	if rValue.Kind() == reflect.Ptr {
		rValue = rValue.Elem()
	}
	for i := 0; i < rFieldNum; i++ {
		rVFi := rValue.Field(i)
		rTFi := rType.Field(i)
		if rVFi.Kind() == reflect.Ptr {
			rVFi = rVFi.Elem()
		}
		if rVFi.CanSet() {
			thisMeta := metaHolder[i]
			electedVal, electionError := thisMeta.electValue()
			if electionError != nil {
				if *electionError == *models.OptionalValueMissing() {
					continue
				}
				return *electionError.Field(rTFi.Name)
			}
			switch rTFi.Type.Kind() {
			case reflect.Int:
				intVal, convErr := strconv.Atoi(electedVal)
				if convErr != nil {
					return *models.InvalidValue().Field(rTFi.Name)
				}
				rVFi.Set(reflect.ValueOf(intVal))
			case reflect.String:
				rVFi.Set(reflect.ValueOf(electedVal))
			}
		}
	}
	return nil
}

type fieldMeta struct {
	position    int
	envName     string
	mandatory   bool
	defVal      string
	valueHolder string
}

//electValue chooses the best value for the variable in hand.
//The preferences goes as
//	1.	The value read from the environment
//	2.	The default value
//	3.	Corresponding zero value if 1 & 2 fails and the variable
//		is set as optional
//Returns error if the value is missing(both in environment
//and as default) and the variable is set to mandatory.
func (f *fieldMeta) electValue() (string, *models.EnvError) {
	if f.valueHolder != "" {
		return f.valueHolder, nil
	} else if f.defVal != "" {
		return f.defVal, nil
	} else if f.mandatory {
		return "", models.MandatoryValueMissing()
	}
	return "", models.OptionalValueMissing()
}

type tagInfo struct {
	name      string
	mandatory bool
	deftVal   string
}

//tag parser performs the parsing of the tag.
//Tag assumes the syntax :
//		`env:"<env_name>;<mandatory/optional>;<default_value>'
//	1. use ';' to separate the elements of the tag;
//	2. use adjacent ';' to represent missing elements
func tagParser(tag string) (tagInfo, *models.EnvError) {
	result := tagInfo{}
	if len(tag) == 0 {
		return result, nil
	}
	exploded := strings.Split(tag, ";")
	explodedLength := len(exploded)
	if explodedLength < 1 {
		return result, models.InvalidTag()
	}

	//1. name
	result.name = strings.TrimSpace(exploded[0])
	if result.name == "" {
		return result, models.InvalidTag()
	}
	if explodedLength < 2 {
		result.mandatory = false
		return result, nil
	}

	//2. optional / mandatory
	if exploded[1] == OPTIONAL {
		result.mandatory = false
	} else if exploded[1] == MANDATORY {
		result.mandatory = true
	} else {
		return result, models.InvalidTag()
	}
	if explodedLength < 3 {
		return result, nil
	}
	//3. default value
	result.deftVal = exploded[2]
	return result, nil
}

func (e EnvToStruct) envToHolder(prefix string, appEnvs map[int]fieldMeta) map[int]fieldMeta {
	envVars := os.Environ()
	//lookUp table construction
	lookUpTable := make(map[string]string)
	for _, envVal := range envVars {
		exploded := strings.Split(envVal, "=")
		lookUpTable[exploded[0]] = exploded[1]
	}
	//cross check with app env
	for i, fMeta := range appEnvs {
		envVal, found := lookUpTable[prefix + fMeta.envName]
		if found {
			envVal = envVal
			i = i
			metaH := fieldMeta{}
			metaH.envName = fMeta.envName
			metaH.defVal = fMeta.defVal
			metaH.valueHolder = envVal
			metaH.mandatory = fMeta.mandatory
			metaH.position = fMeta.position
			appEnvs[i] = metaH
		}
	}
	return appEnvs
}
