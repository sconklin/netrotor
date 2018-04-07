/*
 * N1MM broadcasts Rotator commands from port 12040
 * Rotator status we send are sent from port 13010
 * https://stackoverflow.com/questions/16465705/how-to-handle-configuration-in-go
 */

package main
 
import (
    "encoding/json"
    "fmt"
    "net"
    "os"
)

type Configuration struct {
    Rotators []map[string]string
    AnotherThing string
}

func readConfig(jsonFileName string, rotorData map[string]map[string]string, otherData map[string]string) {

file, err := os.Open(jsonFileName)
    if err != nil {
        fmt.Printf("Unable to open config file: <%s>\n", err)
        os.Exit(1)
    }

    defer file.Close()
    decoder := json.NewDecoder(file)
    configuration := Configuration{}
    err = decoder.Decode(&configuration)
    if err != nil {
        fmt.Println("error1:", err)
        os.Exit(1)
    }

    for _, re := range configuration.Rotators {
        var rname string
        
        if val, ok := re["name"]; ok {
           rname = val
         } else {
           fmt.Printf("Rotator config found with no name defined: %v\n", re)
           os.Exit(1)
         }
         // TODO if we have restirctions like no spaces in names, enforce it here
         fmt.Printf("Rotor name: %s\n", rname)

         // Here I'll read the rest of the config parts, and put them in the rotorData map for use in main()
    }

    fmt.Printf("%s\n", configuration.AnotherThing)
}

/* A Simple function to verify error */
func CheckError(err error) {
    if err  != nil {
        fmt.Println("Error: " , err)
        os.Exit(0)
    }
}

func main() {
     var rotors map[string]map[string]string
     var otherConfig map[string]string

     readConfig("multirotorconf.json", rotors, otherConfig)

     // ignore the rest for now
     os.Exit(0)

    /* Lets prepare a address to listen from any address sending at port 12040*/
    ServerAddr,err := net.ResolveUDPAddr("udp",":12040")
    CheckError(err)
 
    /* Now listen at selected port */
    ServerConn, err := net.ListenUDP("udp", ServerAddr)
    CheckError(err)
    defer ServerConn.Close()
 
    buf := make([]byte, 1024)
 
    for {
        n,addr,err := ServerConn.ReadFromUDP(buf)
        fmt.Println("Received ",string(buf[0:n]), " from ",addr)
 
        if err != nil {
            fmt.Println("Error: ",err)
        } 
    }
}
