package main

import (
    "bufio"
    "fmt"
    "net"
    "time"
)

func buildreq (SESSION string,USERID string,COMMAND string,ARG1 string) string {
    var HEAD string = ConcatStr("","<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?><BroadsoftDocument protocol = \"OCI\" xmlns=\"C\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"><sessionId xmlns=\"\">",SESSION,"</sessionId>")
    var REQ string
    if COMMAND=="AuthenticationRequest" {
        REQ = ConcatStr("","<command xsi:type=\"",COMMAND,"\" xmlns=\"\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"><userId>",USERID,"</userId></command></BroadsoftDocument>")
    }
    if COMMAND=="LoginRequest14sp4" {
        REQ = ConcatStr("","<command xsi:type=\"",COMMAND,"\" xmlns=\"\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"><userId>",USERID,"</userId><signedPassword>",ARG1,"</signedPassword></command></BroadsoftDocument>")
    }
    if COMMAND=="UserCallCenterModifyRequest19" {
        REQ = ConcatStr("","<command xsi:type=\"",COMMAND,"\" xmlns=\"\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"><userId>",USERID,"</userId><agentACDState>",ARG1,"</agentACDState></command></BroadsoftDocument>")
    }
    if COMMAND=="LogoutRequest" {
        REQ = ConcatStr("","<command xsi:type=\"",COMMAND,"\" xmlns=\"\" xmlns:xsi=\"http://www.w3.org/2001/XMLSchema-instance\"><userId>",USERID,"</userId></command></BroadsoftDocument>")
    }
    return ConcatStr("",HEAD,REQ)
}

func OCIPsend(Config ConfigT,owner string,ARG string){
    var SESSION string = randSeq(10)
    var dialer net.Dialer
    dialer.Timeout=time.Second
    chandesc, err := dialer.Dial("tcp",ConcatStr(":",Config.Main.Host,Config.Main.OCIPPort))
    if err != nil {
        LogErr(err,"ocip dial")
        return
    }
    chandesc.SetReadDeadline(time.Now().Add(time.Second))
    REQ := buildreq (SESSION,Config.Main.User,"AuthenticationRequest","")
    fmt.Fprintf(chandesc,"%s",REQ)
    chanreader := bufio.NewReader(chandesc)
    status,err := chanreader.ReadString('\n')
    status,err = chanreader.ReadString('\n')
        Log2Out(status)
    ocip := ParseOCIP([]byte(status))
    resp := MakeDigest(Config.Main.Password,ocip.Nonce)

    REQ=buildreq (SESSION,Config.Main.User,"LoginRequest14sp4",resp)
    fmt.Fprintf(chandesc,"%s",REQ)
    status,err = chanreader.ReadString('\n')
    status,err = chanreader.ReadString('\n')
        Log2Out(status)

    REQ=buildreq (SESSION,owner,"UserCallCenterModifyRequest19",ARG)
    fmt.Fprintf(chandesc,"%s",REQ)
    status,err = chanreader.ReadString('\n')
    status,err = chanreader.ReadString('\n')
        Log2Out(status)

    REQ=buildreq (SESSION,Config.Main.User,"LogoutRequest","")
    fmt.Fprintf(chandesc,"%s",REQ)
    chandesc.Close()
}
