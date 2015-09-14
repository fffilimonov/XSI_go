package main

import (
    "encoding/xml"
)

type EventID struct {
    XMLName xml.Name `xml:"Event"`
    Content string `xml:"eventID"`
}

type ChannelID struct {
    XMLName xml.Name `xml:"Channel"`
    Content string `xml:"channelId"`
}

type CallInfo struct {
    Target string `xml:"targetId"`
    Pers string `xml:"eventData>call>personality"`
    State string `xml:"eventData>call>state"`
    Hook string `xml:"eventData>hookStatus"`
    Addr string `xml:"eventData>call>remoteParty>address"`
    CCstatus string `xml:"eventData>stateInfo>state"`
    CCstatuschanged string `xml:"eventData>agentStateInfo>state"`
}

func ParseData (data []byte) CallInfo {
    var callinfo CallInfo
    xml.Unmarshal(data, &callinfo)
    return callinfo
}

func GetChanID (data []byte) string {
    var channelid ChannelID
    xml.Unmarshal(data, &channelid)
    return channelid.Content
}

func GetEventID (data []byte) string {
    var eventid EventID
    xml.Unmarshal(data, &eventid)
    return eventid.Content
}
