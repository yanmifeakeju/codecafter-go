package main

func main() {
	if err := run(config{}); err != nil {
		panic(err)
	}
}
