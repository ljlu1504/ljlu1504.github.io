package main

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
)

func init() {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath("/etc/appname/")   //
	viper.AddConfigPath("$HOME/.appname")  // call multiple times to add many search paths
	viper.AddConfigPath(".")    // optionally look for config in the working directory
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {
		log.Fatal(err)// Handle errors reading the config file
	}
}

func main()  {
	
	fmt.Println(viper.GetString(`app.name`))
	fmt.Println(viper.GetInt(`app.foo`))
	fmt.Println(viper.GetBool(`app.bar`))
	fmt.Println(viper.GetStringMapString(`app`))
	
}