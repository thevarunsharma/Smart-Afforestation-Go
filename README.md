# Smart Afforestation: Go
Golang implementaion of Smart Afforestation AI agent ([smart-afforestation-python](https://github.com/thevarunsharma/Smart-Afforestation)).

## Usage
- Compile to executable binary as:
```
$ go build smart_afforestation.go
```

- Run the executable binary with command line arguments as:
```
$ ./smart_afforestation AQI AREA BUDGET POPULATION RUNTIME
```
where:
	- AQI: Air Quality Index
	- AREA: Area available for plantation (in sq. metres)
	- BUDGET: Budget for plantation (in Rs.)
	- POPULATION: Population of the Area
	- RUNTIME: Desired running time for GA search
