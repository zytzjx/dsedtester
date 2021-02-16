package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"sync"
)

type diskinfos struct {
	infos       map[string]string
	keymap      map[string]string
	mu          *sync.Mutex
	logfile     *os.File
	errList     map[string]int
	supportList []string
}

func (di *diskinfos) LoadConfig(logf *os.File) error {
	f, err := os.Open("translatefield.json")
	if err != nil {
		return err
	}
	defer f.Close()
	// read our opened jsonFile as a byte array.
	byteValue, _ := ioutil.ReadAll(f)
	// we unmarshal our byteArray which contains our

	err = json.Unmarshal(byteValue, &di.keymap)
	if err != nil {
		return err
	}
	di.infos = make(map[string]string)
	di.errList = make(map[string]int)
	di.mu = &sync.Mutex{}
	di.logfile = logf
	return err
}

func (di *diskinfos) AddSupportList(supp string) {
	if supp != "" {
		di.supportList = append(di.supportList, supp)
	}
}

func (di *diskinfos) AddErrorcodes(sMod string, errcode int) {
	if sMod != "" {
		di.errList[sMod] = errcode
	}
}

func (di *diskinfos) AddInfo2Map(key, value string) error {
	if len(di.keymap) == 0 {
		return errors.New("not found key map")
	}
	if key == "" {
		return errors.New("key is empty")
	}
	kk, OK := di.keymap[key]
	if !OK {
		return errors.New("ingore this key")
	}
	di.mu.Lock()
	defer di.mu.Unlock()
	di.infos[kk] = value
	return nil
}

func (di *diskinfos) WriteLog(s string) {
	if di.logfile == nil {
		return
	}
	di.logfile.WriteString(s)
}

func (di *diskinfos) StoreDB(label int) {
	for k, v := range di.infos {
		if err := Set(label, k, v, 0); err != nil {
			if di.logfile != nil {
				di.WriteLog(err.Error())
			}
		}
	}
}
