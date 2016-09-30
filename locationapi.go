package main

import (
    "net/http"
    "fmt"
    "encoding/json"
    "log"
    "io/ioutil"
    "strconv"
    "cloud.google.com/go/pubsub"
    "golang.org/x/net/context"
    "github.com/tecbot/gorocksdb"
    //"reflect"
    "time"
)
var (
    topic *pubsub.Topic
    db *gorocksdb.DB
    ro *gorocksdb.ReadOptions
    wo *gorocksdb.WriteOptions
)
type LocationFromIP struct{
    Ip              string
    Country_code    string
    Country_name    string
    Region_code     string
    Region_name     string
    City            string
    Zip_code        string
    Time_zone       string
    Latitude        string
    Longitude       string
    Metro_code      string
}

type RerquiredInfo struct{
    Country         string
    Country_code    string
    Region          string
    Region_code     string
    Zip_code        string         
}

type InputData struct{
    Ip          string
    Latitude    float64
    Longitude   float64
}
type GapiAddressComp struct{
    Long_name   string
    Short_name  string
    Types       []string
}
type GapiLocationObject struct{
    Address_components  []GapiAddressComp
    Formatted_address   string
    Geometry            GapiLocationGeometry
    Types               []string
}
type GapiLocationGeometry struct{
    Location        GapiLocation
    Location_type   string
    Viewport        GapiViewReport
    place_id        string       
} 
type GapiLocation struct{
    Lat             float64
    Lng             float64
}
type GapiViewReport struct{
    Northeast       GapiLocation
    Southwest       GapiLocation
}
type GapiLocationResponse struct{
    Results         []GapiLocationObject
    Status          string
}

func FloatToString(val float64) string {
    return strconv.FormatFloat(val, 'f', 7, 64)
}

func IntToString(val int64) string{
    return strconv.FormatInt(val, 10)
}

func APIHandler(response http.ResponseWriter, req *http.Request){
    
    var reqInfo RerquiredInfo
    requestbody, _ := ioutil.ReadAll(req.Body)
    log.Println("Request Body:", string(requestbody))
    
    var input InputData
    err := json.Unmarshal(requestbody, &input)
    if err == nil{
        log.Println("Afetr unmarshal")
        log.Println("IP: ", input.Ip)
        log.Println("Latitude: ", input.Latitude)
        log.Println("Longitude: ", input.Longitude)
    }else{
        log.Println(err.Error())
    }

    if input.Ip == "" {
        //get From co-ordinates
        url := "https://maps.googleapis.com/maps/api/geocode/json?latlng="+FloatToString(input.Latitude)+","+FloatToString(input.Longitude)+"&key=AIzaSyDbaAWkQs-cESgDdl02Q6l0TpfA4IBpw8I"
        gapiResponse, err := http.Get(url)
        if err != nil {
            panic(err.Error())
        }
        body, err := ioutil.ReadAll(gapiResponse.Body)
        if err != nil {
            panic(err.Error())
        }
        var gapiLocation GapiLocationResponse
        err = json.Unmarshal(body, &gapiLocation)
        if err != nil {
            panic(err.Error())
        }
        fmt.Printf(gapiLocation.Status)
        log.Println(gapiLocation.Results[0].Formatted_address)

        for _, addComp := range gapiLocation.Results[0].Address_components{

            switch addComp.Types[0] {
                case "postal_code":
                    reqInfo.Zip_code = addComp.Long_name
                case "country":
                    reqInfo.Country = addComp.Long_name
                    reqInfo.Country_code = addComp.Short_name

                case "administrative_area_level_1":
                    reqInfo.Region = addComp.Long_name
                    reqInfo.Region_code  = addComp.Short_name
                default:
            }  
        }

    }else{
        //get from IP
        url := "https://freegeoip.net/json/"+ input.Ip
        log.Println(url)
        geoipResponse, err3 := http.Get(url)
        if err3 != nil {
            log.Println(err3.Error())
        }
        defer geoipResponse.Body.Close()

        decoder := json.NewDecoder(geoipResponse.Body)
        var locFromIp LocationFromIP
        err = decoder.Decode(&locFromIp)
        reqInfo.Zip_code = locFromIp.Zip_code
        reqInfo.Country = locFromIp.Country_name
        reqInfo.Country_code = locFromIp.Country_code
        reqInfo.Region = locFromIp.Region_name
        reqInfo.Region_code = locFromIp.Region_code
        
    }

    go publishUpdate(&reqInfo)
    go writeToDB(&reqInfo)
    fmt.Fprintf(response, "You are in- " + reqInfo.Region +", "+ reqInfo.Region_code +", "+ reqInfo.Country+", "+ reqInfo.Country_code +", "+ reqInfo.Zip_code);

    // value, err := db.Get(ro, []byte("key2"))
    // defer value.Free()
    // defer db.Close();
    // log.Println("Fetched from database") 
    // log.Println(string(value.Data()[:]))
    // log.Println(reflect.TypeOf(value))
}


func main() {
    //Creating DB instance - START
    var err error
    bbto := gorocksdb.NewDefaultBlockBasedTableOptions()
    bbto.SetBlockCache(gorocksdb.NewLRUCache(3 << 30))
    opts := gorocksdb.NewDefaultOptions()
    opts.SetBlockBasedTableFactory(bbto)
    opts.SetCreateIfMissing(true)
    db, err = gorocksdb.OpenDb(opts, "newdb") 

    ro = gorocksdb.NewDefaultReadOptions()
    ro.SetFillCache(false)
    wo = gorocksdb.NewDefaultWriteOptions()
    defer db.Close()

    // if ro and wo are not used again, be sure to Close them.
    //err = db.Put(wo, []byte("key2"), []byte("Shashwat2"))
    


    //Creating DB instance - END

    //Creating client and topic - START
    ctx := context.Background()
    client, err := pubsub.NewClient(ctx, "pretlist-daemons-apps-us-east1")
    if err != nil {
        log.Fatalf("Could not create pubsub Client: %v", err)
    }
    if client!=nil {
        log.Println("Client found");
    }

    fmt.Println("Listing all topics from the project:")
    topics, err := list(client)
    if err != nil {
        log.Fatalf("Failed to list topics: %v", err)
    }
    for _, t := range topics {
        fmt.Println(t)
    }

    const topicname = "loc-service"
    topic, _ = client.CreateTopic(ctx, topicname)
    //Creating client and topic - END

    http.HandleFunc("/", static)
    http.HandleFunc("/api/", APIHandler)
    err = http.ListenAndServe(":8080", nil)
    if err != nil {
        log.Println("ListenAndServe: "+err.Error())
    }
    log.Println("Listening..")
    
}

func static(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "files/"+r.URL.Path)
}

func list(client *pubsub.Client) ([]*pubsub.Topic, error) {
    ctx := context.Background()
    
    // [START list_topics]
    var topics []*pubsub.Topic

    it := client.Topics(ctx)
    for {
        topic, err := it.Next()
        if err == pubsub.Done {
            break
        }
        if err != nil {
            return nil, err
        }
        topics = append(topics, topic)
    }

    return topics, nil
    // [END list_topics]
}

func create(client *pubsub.Client, topic string) error {
    ctx := context.Background()
    // [START create_topic]
    t, err := client.CreateTopic(ctx, topic)
    if err != nil {
        return err
    }
    fmt.Printf("Topic created: %v\n", t)
    // [END create_topic]
    return nil
}

func publishUpdate(locInfo *RerquiredInfo) {
   
    ctx := context.Background()

    b, err := json.Marshal(locInfo)
    if err != nil {
        log.Println(err.Error())
        return
    }
    _, err = topic.Publish(ctx, &pubsub.Message{Data: b})
    log.Println("Published update to Pub/Sub for Book ID %d: %v", locInfo, err)
}
func writeToDB(locInfo *RerquiredInfo) {

    //Testing time
    key := IntToString(time.Now().UnixNano() / int64(time.Millisecond))
    
    fmt.Println("key: ", key)

    b, err := json.Marshal(locInfo)
    if err != nil {
        log.Println(err.Error())
        return
    }
   err = db.Put(wo,[]byte(key), []byte(b))
   if err == nil {
        fmt.Println("Written to DB")
   }
   readSavedInfo()

}
func readSavedInfo(){
    it := db.NewIterator(ro)
    defer it.Close()
    it.SeekToFirst()
    count := 1
   // it.Seek([]byte("foo"))
    for it = it; it.Valid(); it.Next() {
        key := it.Key()
        value := it.Value()
        fmt.Printf("%v. Key: %v - Value: %v\n", count, string(key.Data()[:]), string(value.Data()[:]))
        key.Free()
        value.Free()
        count++
    }
    if err := it.Err(); err != nil {
        fmt.Println("Issue with iterator")
    }
}
