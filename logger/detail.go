package logger

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"time"
)

type DetailLog interface {
	IsRawDataEnabled() bool
	AddInputRequest(node, cmd, invoke string, rawData, data interface{})
	AddInputHttpRequest(node, cmd, invoke string, req *http.Request, rawData bool)
	AddOutputRequest(node, cmd, invoke string, rawData, data interface{})
	End()
	AddInputResponse(node, cmd, invoke string, rawData, data interface{}, protocol, protocolMethod string)
	AddOutputResponse(node, cmd, invoke string, rawData, data interface{})
	AutoEnd() bool
}

func NewDetailLog(Session, initInvoke, scenario string) DetailLog {
	// session := req.Context().Value(xSession)
	currentTime := time.Now()
	if Session == "" {
		Session = fmt.Sprintf("default_%s", currentTime.Format("20060102150405"))
	}

	if initInvoke == "" {
		initInvoke = fmt.Sprintf("%s_%s", configLog.ProjectName, currentTime.Format("20060102150405"))
	}
	host, _ := os.Hostname()
	data := &detailLog{
		LogType:       "Detail",
		Host:          host,
		AppName:       configLog.ProjectName,
		Instance:      getInstance(),
		Session:       Session,
		InitInvoke:    initInvoke,
		Scenario:      scenario,
		Input:         []InputOutputLog{},
		Output:        []InputOutputLog{},
		conf:          configLog.Detail,
		startTimeDate: time.Now(),
		timeCounter:   make(map[string]time.Time),
		// req:           req,
	}

	return data
}

func getInstance() *string {
	instance := fmt.Sprintf("%d", os.Getpid())
	return &instance
}

func (dl *detailLog) IsRawDataEnabled() bool {
	return dl.conf.RawData
}

type InComing struct {
	Header any        `json:"header,omitempty"`
	Query  url.Values `json:"query,omitempty"`
	Body   any        `json:"body,omitempty"`
}

func (dl *detailLog) AddInputHttpRequest(node, cmd, invoke string, req *http.Request, rawData bool) {
	data := InComing{
		Header: req.Header,
		Query:  req.URL.Query(),
		Body:   nil,
	}

	bodyBytes, _ := io.ReadAll(req.Body)

	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	json.Unmarshal(bodyBytes, &data.Body)

	var raw string
	if rawData {
		raw = ToJson(data)
	}

	protocol := req.Proto
	protocolMethod := req.Method
	dl.addInput(&logEvent{
		node:           node,
		cmd:            cmd,
		invoke:         invoke,
		logType:        "req",
		rawData:        raw,
		data:           data,
		protocol:       protocol,
		protocolMethod: protocolMethod,
	})
}

func (dl *detailLog) AddInputRequest(node, cmd, invoke string, rawData, data interface{}) {
	if rawData != nil {
		if _, ok := rawData.(string); !ok {
			rawData = ToJson(rawData)
		}
	}
	dl.addInput(&logEvent{
		node:    node,
		cmd:     cmd,
		invoke:  invoke,
		logType: "req",
		rawData: rawData,
		data:    data,
		// protocol:       dl.req.Proto,
		// protocolMethod: dl.req.Method,
	})
}

func (dl *detailLog) AddInputResponse(node, cmd, invoke string, rawData, data interface{}, protocol, protocolMethod string) {
	resTime := time.Now().Format(time.RFC3339)
	if rawData != nil {
		if _, ok := rawData.(string); !ok {
			rawData = ToJson(rawData)
		}
	}
	dl.addInput(&logEvent{
		node:           node,
		cmd:            cmd,
		invoke:         invoke,
		logType:        "res",
		rawData:        rawData,
		data:           ToStruct(data),
		resTime:        resTime,
		protocol:       protocol,
		protocolMethod: protocolMethod,
	})
}

func (dl *detailLog) AddOutputResponse(node, cmd, invoke string, rawData, data interface{}) {
	if rawData != nil {
		if _, ok := rawData.(string); !ok {
			// rawData = fmt.Sprintf("%v", rawData)
			rawData = ToJson(rawData)
		}
	}
	dl.AddOutput(logEvent{
		node:    node,
		cmd:     cmd,
		invoke:  invoke,
		logType: "res",
		rawData: rawData,
		data:    ToStruct(data),
	})
	// dl.End()
}

func (dl *detailLog) addInput(input *logEvent) {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	now := time.Now()
	if dl.startTimeDate.IsZero() {
		dl.startTimeDate = now
	}

	var resTimeString string
	if input.resTime != "" {
		resTimeString = input.resTime
	} else if input.logType == "res" {
		if startTime, exists := dl.timeCounter[input.invoke]; exists {
			duration := time.Since(startTime).Milliseconds()
			resTimeString = fmt.Sprintf("%d ms", duration)
			delete(dl.timeCounter, input.invoke)
		}
	}

	protocolValue := dl.buildValueProtocol(&input.protocol, &input.protocolMethod)
	inputLog := InputOutputLog{
		Invoke:   input.invoke,
		Event:    fmt.Sprintf("%s.%s", input.node, input.cmd),
		Protocol: protocolValue,
		Type:     input.logType,
		RawData:  dl.isRawDataEnabledIf(input.rawData),
		Data:     input.data,
		ResTime:  &resTimeString,
	}
	dl.Input = append(dl.Input, inputLog)
}

func (dl *detailLog) AddOutputRequest(node, cmd, invoke string, rawData, data interface{}) {
	if rawData != nil {
		if _, ok := rawData.(string); !ok {
			rawData = ToJson(rawData)
		}
	}
	dl.AddOutput(logEvent{
		node:    node,
		cmd:     cmd,
		invoke:  invoke,
		logType: "rep",
		rawData: rawData,
		data:    ToStruct(data),
		// protocol:       dl.req.Proto,
		// protocolMethod: dl.req.Method,
	})
}

func (dl *detailLog) AddOutput(out logEvent) {
	dl.mu.Lock()
	defer dl.mu.Unlock()
	now := time.Now()
	if out.invoke != "" && out.logType != "res" {
		dl.timeCounter[out.invoke] = now
	}

	protocolValue := dl.buildValueProtocol(&out.protocol, &out.protocolMethod)
	outputLog := InputOutputLog{
		Invoke:   out.invoke,
		Event:    fmt.Sprintf("%s.%s", out.node, out.cmd),
		Protocol: protocolValue,
		Type:     out.logType,
		RawData:  dl.isRawDataEnabledIf(out.rawData),
		Data:     out.data,
	}
	dl.Output = append(dl.Output, outputLog)
}

func (dl *detailLog) End() {
	if dl.startTimeDate.IsZero() {
		log.Fatal("end() called without any input/output")
	}

	processingTime := fmt.Sprintf("%d ms", time.Since(dl.startTimeDate).Milliseconds())
	dl.ProcessingTime = &processingTime

	inputTimeStamp := dl.formatTime(dl.inputTime)
	dl.InputTimeStamp = inputTimeStamp

	outputTimeStamp := dl.formatTime(dl.outputTime)
	dl.OutputTimeStamp = outputTimeStamp

	logDetail, _ := json.Marshal(dl)
	if dl.conf.LogConsole {
		os.Stdout.Write(logDetail)
		os.Stdout.Write([]byte(endOfLine()))
	}

	if dl.conf.LogFile && dl.conf.LogDetail != nil {
		dl.conf.LogDetail.Info(string(logDetail))
	}

	dl.clear()
}

func (dl *detailLog) buildValueProtocol(protocol, method *string) *string {
	if protocol == nil {
		return nil
	}
	result := *protocol
	if method != nil {
		result += "." + *method
	}
	return &result
}

func (dl *detailLog) AutoEnd() bool {
	if dl.startTimeDate.IsZero() {
		return false
	}
	if len(dl.Input) == 0 && len(dl.Output) == 0 {
		return false
	}

	dl.End()
	return true
}

func (dl *detailLog) isRawDataEnabledIf(rawData interface{}) interface{} {
	if dl.conf.RawData {
		return rawData
	}
	return nil
}

func (dl *detailLog) formatTime(t *time.Time) *string {
	if t == nil {
		return nil
	}
	ts := t.Format(time.RFC3339)
	return &ts
}

func endOfLine() string {
	if runtime.GOOS == "windows" {
		return "\r\n"
	}
	return "\n"
}

func (dl *detailLog) clear() {
	dl.ProcessingTime = nil
	dl.InputTimeStamp = nil
	dl.OutputTimeStamp = nil
	dl.Input = nil
	dl.Output = nil
	dl.startTimeDate = time.Time{}
}

func ToJson(data interface{}) string {
	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error: ", err)
		return fmt.Sprintf("%v", data)
	}
	return string(jsonData)
}

// convert struct to json
func ToStruct(data interface{}) (result interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return data
	}

	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return data
	}
	return result
}
