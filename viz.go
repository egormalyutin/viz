package main

func main() {
	Provider.Set("csv")
	Provider.Init("data.csv")
	Serve()
}
