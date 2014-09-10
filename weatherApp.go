package weatherApp

import (
    "fmt"
    "html/template"
    "net/http"
    "encoding/json"
    "appengine"
    "appengine/urlfetch"
    "log"
    "errors"
    
   
)

func init() {
    http.HandleFunc("/", root)
    http.HandleFunc("/temperature", temperature)
}

func root(w http.ResponseWriter, r *http.Request) {

    fmt.Fprint(w, inputForm)
}

const inputForm = `
<html>
  <body>
    <h1><i>Enter names of the cities for Today's Weather:<i></h1>
    <form action="/temperature" method="post">
    <h4>City1 :</h4>
      <div><textarea name="city1" rows="3" cols="60"></textarea></div>
      <h4>City2 :</h4>
      <div><textarea name="city2" rows="3" cols="60"></textarea></div>
      <h4>City3 :</h4>
      <div><textarea name="city3" rows="3" cols="60"></textarea></div>
      <h4>City4 :</h4>
      <div><textarea name="city4" rows="3" cols="60"></textarea></div>
      <h4>City5 :</h4>
      <div><textarea name="city5" rows="3" cols="60"></textarea></div>
      <br><br>
      <div><input type="submit" value="Submit" style="height:1000px; width:250px"></div>
    </form>
  </body>
</html>
`

func temperature(w http.ResponseWriter, r *http.Request) {

    

    todaysWeatherReports := make(chan todaysWeather)
    errs := make(chan error)

    
    var cities [5] string
    cities[0]=r.FormValue("city1")
    cities[1]=r.FormValue("city2")
    cities[2]=r.FormValue("city3")
    cities[3]=r.FormValue("city4")
    cities[4]=r.FormValue("city5")

    
    for _, city := range cities {
       go func(c string) {



            data, fetcherr := query(c, r)


            if fetcherr != nil {
                errs <- fetcherr
                return
            }


            var weather todaysWeather

            weather.MaximumTemperature = data.Data.Weather[0].Max
            weather.MinimumTemperature = data.Data.Weather[0].Min
            weather.Name = data.Data.Request[0].City


            todaysWeatherReports <- weather
        }(city)


    }

    var reports []todaysWeather



    
    for i := 0; i < 5; i++ {
        select {
        case temp := <-todaysWeatherReports:

            reports = append(reports, temp)


        case err := <-errs:
            log.Print("error")
            log.Print(err)

        }
    }


    err := temperatureOutput.Execute(w,reports)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
    }
}

func query(city string, r *http.Request) (todaysWeatherData, error) {
	c := appengine.NewContext(r)
    client := urlfetch.Client(c)




    resp, err := client.Get("http://api.worldweatheronline.com/free/v1/weather.ashx?key=435e115161bbf12322d04cc9cf46379491ca0f3a&format=json&q=" + city)
    if err != nil {
        return todaysWeatherData{}, err
    }

    defer resp.Body.Close()

    var d todaysWeatherData


    if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
        return todaysWeatherData{}, err
    }

    if (len(d.Data.ErrorM) > 0) {
        return todaysWeatherData{}, errors.New(d.Data.ErrorM[0].Message)
    }


    return d, nil
}




type Data1 struct {
    Weather []weather `json:"weather"`
    Request []request `json:"request"`
    ErrorM []errorm `json:"error"`


}

type weather struct {
    Max string `json:"tempMaxC"`
    Min string `json:"tempMinC"`


}

type request struct {
    City string `json:"query"`

}


type errorm struct {
    Message string `json:"msg"`

}


type todaysWeatherData struct {
    Data Data1 `json:"data"`
}


type todaysWeather struct {
    MaximumTemperature string
    MinimumTemperature string
    Name string
}





var temperatureOutput = template.Must(template.New("temperature").Parse(temperatureOutputHTML))

const temperatureOutputHTML = `
<html>
  <body>

    {{ $cities := . }} 
    {{range $index, $element := $cities }}
       <h1><i>{{$element.Name}}</i></h1><br><h3>Maximum Temperature: {{$element.MaximumTemperature}} C</h3><h3>Minimum Temperature: {{$element.MinimumTemperature}} C </h3><br><br>
    {{end}}


  </body>
</html>
`