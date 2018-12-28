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
	NameLong            string `json:"name_long"`
	SoftwareVersion     string `json:"software_version"`
	StructureID         string `json:"structure_id"`
}

type NestmonConfig struct {
	AccessToken       string `json:"AccessToken"`
	DbHostUrl         string `json:"DbHostUrl"`
	NestDbName        string `json:"NestDbName"`
	NestDbUsername    string `json:"NestDbUsername"`
	NestDbPassword    string `json:"NestDbPassword"`
	WeatherDbName     string `json:"WeatherDbName"`
	WeatherDbUsername string `json:"WeatherDbUsername"`
	WeatherDbPassword string `json:"WeatherDbPassword"`
	OWMZip            string `json:"OWMZip"`
	OWMAppId          string `json:"OWMAppId"`
}

// Weather data from openweathermap.org
type OWMResponse struct {
	Main *OMWResponseMain `json:"main"`
	Wind *OMWResponseWind `json:"wind"`
}

type OMWResponseMain struct {
	Temp     float32 `json:"temp"`
	Humidity float32 `json:"humidity"`
}

type OMWResponseWind struct {
	Speed float32 `json:"speed"`
}
