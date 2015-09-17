package main

import (
    "bufio"
    "fmt"
    "net"
    "os"
    "strconv"
    "strings"
    "time"
)

type DefHead struct {
    AUTHORIZATION string
    HOSTH string
    CTYPE string
    CHANID string
}

func MakeDef (Config ConfigT) DefHead {
    var def DefHead
    var base64 string = MakeAuth(Config.Main.User,Config.Main.Password)
    def.AUTHORIZATION = ConcatStr("","Authorization: Basic ",base64)
    def.HOSTH = ConcatStr("","Host: ",Config.Main.HTTPHost)
    def.CTYPE = "Content-Type: application/x-www-form-urlencoded"
    def.CHANID = randSeq(10);
    return def
}

func XSISubscribeCH (Config ConfigT,def DefHead) (net.Conn,string) {
    var CPOST string = "POST /com.broadsoft.async/com.broadsoft.xsi-events/v2.0/channel HTTP/1.1";
    var CSET string = ConcatStr("","<?xml version=\"1.0\" encoding=\"UTF-8\"?><Channel xmlns=\"http://schema.broadsoft.com/xsi\"><channelSetId>",def.CHANID,"</channelSetId><priority>1</priority><weight>100</weight><expires>",Config.Main.Expires,"</expires></Channel>")
    var CLEN string = ConcatStr("","Content-Length: ",strconv.Itoa(len(CSET)))
    var dialer net.Dialer
    dialer.Timeout=time.Second
    chandesc, err := dialer.Dial("tcp", ConcatStr(":",Config.Main.Host,Config.Main.HTTPPort))
    if err != nil {
        LogErr(err,"chan dial")
    }
    fmt.Fprintf(chandesc,"%s\n%s\n%s\n%s\n%s\n\n%s\n",CPOST,def.AUTHORIZATION,def.HOSTH,CLEN,def.CTYPE,CSET)
    chanreader := bufio.NewReader(chandesc)
    status,err := chanreader.ReadString('\n')
    if err != nil {
        LogErr(err,"chan read")
    }
    if !strings.Contains(status,"200") {
        LogErr(nil,"chan status",status)
    }
//get new chan id
    data := make([]byte, 1024)
    _,err = chanreader.Read(data)
    chanID := GetChanID(data)
    return chandesc,chanID
}

func XSISubscribe (Config ConfigT,def DefHead,target string,event string) {
    var CHANPOST string = ConcatStr("","POST /com.broadsoft.xsi-events/v2.0/",Config.Main.Target,"/",target,"/subscription HTTP/1.1")
    var CHANSET string = ConcatStr("","<Subscription xmlns=\"http://schema.broadsoft.com/xsi\"><event>",event,"</event><expires>",Config.Main.Expires,"</expires><channelSetId>",def.CHANID,"</channelSetId><applicationId>CommPilotApplication</applicationId></Subscription>")
    var CHANLEN string = ConcatStr("","Content-Length: ",strconv.Itoa(len(CHANSET)))
    subdesc, err := net.Dial("tcp", ConcatStr(":",Config.Main.Host,Config.Main.HTTPPort))
    if err != nil {
        LogErr(err,"sub dial")
    }
    fmt.Fprintf(subdesc,"%s\n%s\n%s\n%s\n%s\n\n%s\n",CHANPOST,def.AUTHORIZATION,def.HOSTH,CHANLEN,def.CTYPE,CHANSET)
    subreader := bufio.NewReader(subdesc)
    status,err := subreader.ReadString('\n')
    if err != nil {
        LogErr(err,"sub read")
    }
    if !strings.Contains(status,"200") {
        LogErr(nil,"sub status",status)
    }
    subdesc.Close()
}

func XSIResponse (ID string,def DefHead,Config ConfigT) {
    var status string
    var CONFPOST string = "POST /com.broadsoft.xsi-events/v2.0/channel/eventresponse HTTP/1.1"
    var CONFSET string = ConcatStr("","<?xml version=\"1.0\" encoding=\"UTF-8\"?><EventResponse xmlns=\"http://schema.broadsoft.com/xsi\"><eventID>",ID,"</eventID><statusCode>200</statusCode><reason>OK</reason></EventResponse>")
    var CONFLEN string = ConcatStr("","Content-Length: ",strconv.Itoa(len(CONFSET)))
    respdesc, err := net.Dial("tcp", ConcatStr(":",Config.Main.Host,Config.Main.HTTPPort))
    if err != nil {
        LogErr(err,"resp dial")
        os.Exit(1)
    }
    fmt.Fprintf(respdesc,"%s\n%s\n%s\n%s\n%s\n\n%s\n",CONFPOST,def.AUTHORIZATION,def.HOSTH,CONFLEN,def.CTYPE,CONFSET)
    respreader := bufio.NewReader(respdesc)
    status,err = respreader.ReadString('\n')
    if err != nil {
        LogErr(err,"resp read")
        os.Exit(1)
    }
    if !strings.Contains(status,"200") {
        LogErr(nil,"resp status",ID,status)
    }
    respdesc.Close()
}

func XSIresubscribe(Config ConfigT,cCh chan net.Conn,owner string) {
    exp,_ := strconv.Atoi(Config.Main.Expires)
    timer := time.NewTimer(time.Nanosecond)
    timer2 := time.NewTimer(time.Second)
    var lchanID string
    var nilchannel,lchannel net.Conn
    var def DefHead
    for {
        select {
            case <-timer.C:
                cCh <- nilchannel
                if lchannel != nil {
                    lchannel.Close()
                }
                def = MakeDef(Config)
                channel,chanID := XSISubscribeCH(Config,def)
                lchannel = channel
                lchanID = chanID
                for _,event := range Config.Main.Event {
                    XSISubscribe(Config,def,owner,event)
                    cCh <- channel
                    time.Sleep(time.Millisecond*100)
                }
                XSISubscribe(Config,def,Config.Main.CCID,Config.Main.CCEvent)
                timer.Reset(time.Second*time.Duration(exp))
                timer2.Reset(time.Second*6)
            case <-timer2.C:
                res := XSIheartbeat(Config,def,lchanID)
                if res>0 {
                    timer.Reset(time.Nanosecond)
                }
                timer2.Reset(time.Second*6)
        }
    }
}

func XSIread(ch chan string,cCh chan net.Conn) {
    var lchanel net.Conn
    data := make([]byte, 2048)
    var breader *bufio.Reader
    for {
        select {
            case chanel := <-cCh:
                lchanel = chanel
                breader = bufio.NewReader(lchanel)
            default:
                if lchanel != nil {
                    lchanel.SetReadDeadline(time.Now().Add(time.Second))
                    status,_ := breader.ReadString('\n')
                    inbuf,_ := strconv.ParseInt(strings.Trim(status,"\r\n"),16,32)
                    if inbuf>0 {
                        data,_ = breader.Peek(int(inbuf))
                        ch<- string(data)
                    }
                }
            }
    }
}


func XSIheartbeat(Config ConfigT,def DefHead,channelID string) int {
    var result int = 0
    var PUTHEARTBEAT string =ConcatStr("","PUT /com.broadsoft.xsi-events/v2.0/channel/",channelID,"/heartbeat HTTP/1.1")
    hdesc, err := net.Dial("tcp", ConcatStr(":",Config.Main.Host,Config.Main.HTTPPort))
    if err != nil {
        LogErr(err,"HB dial")
    }
    fmt.Fprintf(hdesc,"%s\n%s\n%s\n\n",PUTHEARTBEAT,def.AUTHORIZATION,def.HOSTH)
    hreader := bufio.NewReader(hdesc)
    status,err := hreader.ReadString('\n')
    if err != nil {
        LogErr(err,"HB read")
    }
    if !strings.Contains(status,"200") {
        LogErr(nil,"HB status",channelID,status)
        result = 1
    }
    hdesc.Close()
    return result
}

func XSImain(Config ConfigT,def DefHead,ch chan string,datach chan string) {
    for {
        select {
            case data := <-ch:
                eventID := GetEventID([]byte(data))
                if eventID != "" {
                    XSIResponse(eventID,def,Config)
                    datach <- data
                    time.Sleep(time.Millisecond*100)
                } else {
                    LogErr(nil,data)
                }
            default:
                time.Sleep(time.Millisecond*10)
        }
    }
}

func XSIGetHook (Config ConfigT,def DefHead,target string) string{
    var GET string = ConcatStr("","GET /com.broadsoft.xsi-actions/v2.0/user/",target,"/calls/HookStatus HTTP/1.1")
    subdesc, err := net.Dial("tcp", ConcatStr(":",Config.Main.Host,Config.Main.HTTPPort))
    if err != nil {
        LogErr(err,"hook dial")
    }
    fmt.Fprintf(subdesc,"%s\n%s\n%s\n\n",GET,def.AUTHORIZATION,def.HOSTH)
    subreader := bufio.NewReader(subdesc)
    status,err := subreader.ReadString('\n')
    if err != nil {
        LogErr(err,"hook read")
    }
    if !strings.Contains(status,"200") {
        LogErr(nil,"hook status",status)
    }
    data := make([]byte, 1024)
    _,err = subreader.Read(data)
    hook := GetHook(data)
    subdesc.Close()
    return hook
}

func XSITransfer (Config ConfigT,def DefHead,target string,CallID string,totarget string) {
    if CallID != "" {
        var PUT string = ConcatStr("","PUT /com.broadsoft.xsi-actions/v2.0/user/",target,"/calls/",CallID,"/BlindTransfer?address=",totarget," HTTP/1.1")
        subdesc, err := net.Dial("tcp", ConcatStr(":",Config.Main.Host,Config.Main.HTTPPort))
        if err != nil {
            LogErr(err,"transfer dial")
        }
        fmt.Fprintf(subdesc,"%s\n%s\n%s\n\n",PUT,def.AUTHORIZATION,def.HOSTH)
        subreader := bufio.NewReader(subdesc)
        status,err := subreader.ReadString('\n')
        if err != nil {
            LogErr(err,"transfer read")
        }
        if !strings.Contains(status,"200") {
            LogErr(nil,"transfer status",status)
        }
        subdesc.Close()
    }
}
