package main
 
import (
    "encoding/json"
    "fmt"
    "net"
    "os"
)
 
/* A Simple function to verify error */
func CheckError(err error) {
    if err  != nil {
        fmt.Println("Error: " , err)
        os.Exit(0)
    }
}

/*
 * The configuration can contain multiple rotators.
 * Each rotator has the following information:
 *   Rotator name
 *   Associated serial port
 *   serial port speed
 *   rotctl model number
 *
 */
 
type Configuration struct {
    Users    []string
    Groups   []string
}


/*
 * N1MM broadcasts Rotator commands from port 12040
 * Rotator status we send are sent from port 13010
 * https://stackoverflow.com/questions/16465705/how-to-handle-configuration-in-go
 */

func main() {
    /* get configuration data */
    file, _ := os.Open("rotorconf.json")
    defer file.Close()
    decoder := json.NewDecoder(file)
    configuration := Configuration{}
    err := decoder.Decode(&configuration)
    if err != nil {
       fmt.Println("error:", err)
    }


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
