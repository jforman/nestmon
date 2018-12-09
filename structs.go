package nestmon

// Root of the NestAPI Response JSON
type NestAPIResponse struct {
	Devices    *Devices              `json:"devices",omitempty`
	Structures map[string]*Structure `json:"structures,omitempty`
}

// Root of NestAPI Streaming JSON Response
type NestAPIStreamingResponse struct {
	Data *NestAPIStreamingData `json:"data"`
}

type NestAPIStreamingData struct {
	Devices    *Devices              `json:"devices",omitempty`
	Structures map[string]*Structure `json:"structures",omitempty`
}

type Devices struct {
	Thermostats map[string]*Thermostat `json:"thermostats",omitempty`
}

type Structure struct {
	Name     string `json:"name",omitempty`
	TimeZone string `json:"time_zone",omitempty`
}

type Thermostat struct {
	AmbientTemperatureF int    `json:"ambient_temperature_f"`
	Humidity            int    `json:"humidity"`
	HvacState           string `json:"hvac_state"`
	Name                string `json:"name"`
	SoftwareVersion     string `json:"software_version"`
}

type NestmonConfig struct {
	AccessToken string `json:"AccessToken"`
}
