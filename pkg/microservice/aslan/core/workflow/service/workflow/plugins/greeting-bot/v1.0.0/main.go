package main

import (
	"fmt"

	"github.com/spf13/viper"
)

const (
	WhoAmI        = "WHO_AM_I"
	WeatherStatus = "WEATHER_STATUE"
)

func main() {
	viper.AutomaticEnv()

	who_am_i := viper.GetString(WhoAmI)
	weather_status := viper.GetString(WeatherStatus)
	fmt.Printf("Hello %s, today is %s", who_am_i, weather_status)
}
