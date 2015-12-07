# env_parser
A parser for parsing environment variables to a struct in golang

# Usage
Get the package first.

**go get github.com/shalinlk/env_parser**

# Sample Program

```sh
type Source struct {
	Name string `env:"name;mandatory"`
	Age  int    `env:"age;optional;25"`
}

func main() {
	ep := env_parser.NewEnvParser()
	//set name : optional
	ep.Name("test_app")
	//set separator : optional
	ep.Separator("_")

	source := Source{}
	er := ep.Map(&source)
	if er != nil {
		fmt.Println("Error : ", er)
		return
	}
	fmt.Println("Final :", source)
}

```
