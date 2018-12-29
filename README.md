# Nestmon

A polling (or streaming) monitor to receive Nest thermostat updates and act on them.

## Executing

### thermostat_status

```
Usage of /tmp/go-build100978462/b001/exe/thermostat_status:
  -config string
    	JSON config containing Nest API access parameters.
  -enableNest
    	Enable checking for Nest data and inserting into database.
  -enableWeather
    	Enable checking the local weather and inserting into database
  -query_interval duration
    	Interval between Nest API queries. (default 3m0s)
```

## Configuration Format

A json-formatted config is used to specify the Access Token:

## Getting the Access Token

```
export AUTH_CODE=PIN
export CLIENT_ID=client-id-from-developers-nest-com
export CLIENT_SECRET=client-secret-from-developers-nest-com
curl -X POST \
  -d "code=$AUTH_CODE&client_id=$CLIENT_ID&client_secret=$CLIENT_SECRET&grant_type=authorization_code" \
  "https://api.home.nest.com/oauth2/access_token"
```
