/*
Copyright 2017 The Camlistore Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Package aboutdialog provides a menu item element that is used in the
// header menu of the web UI, to display an "About" dialog.
package aboutdialog

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"camlistore.org/pkg/auth"

	"honnef.co/go/js/dom"
	"myitcv.io/react"
)

//go:generate reactGen

// New returns the menu item element. It should be used as the entry point, to
// create the needed react element.
//
// config is the web UI config that was fetched from the server.
//
// entryText is the text for this menu item, displayed in the menu.
//
// dialog is the text that appears in the dialog that is created when this menu
// item is clicked on.
//
// class is the CSS class for this item.
func New(entryText, dialog, class string,
	config map[string]string) react.Element {
	if config == nil {
		fmt.Println("Nil config for DownloadItemsBtn")
		return nil
	}
	authToken, ok := config["authToken"]
	if !ok {
		fmt.Println("No authToken in config for AboutMenuItem")
		return nil
	}
	statusRoot, ok := config["statusRoot"]
	if !ok {
		fmt.Println("No statusRoot in config for AboutMenuItem")
		return nil
	}
	props := AboutMenuItemProps{
		class:      class,
		menuEntry:  entryText,
		dialog:     dialog,
		authToken:  authToken,
		statusRoot: statusRoot,
	}
	return buildAboutMenuItemElem(props)
}

type AboutMenuItemDef struct {
	react.ComponentDef
}

type AboutMenuItemProps struct {
	class      string
	menuEntry  string
	dialog     string
	authToken  string
	statusRoot string
}

func (a AboutMenuItemDef) Render() react.Element {
	return react.Div(
		&react.DivProps{
			ClassName: a.Props().class,
			OnClick:   a,
		},
		react.S(a.Props().menuEntry),
	)
}

type status struct {
	Version string // Camlistore build version
	GoInfo  string
}

func (a AboutMenuItemDef) OnClick(e *react.SyntheticMouseEvent) {
	go func() {
		dialogText := a.Props().dialog
		if err := func() error {
			authToken := a.Props().authToken
			am, err := auth.NewTokenAuth(authToken)
			if err != nil {
				return fmt.Errorf("Error setting up auth for download request: %v", err)
			}
			statusPrefix := a.Props().statusRoot
			req, err := http.NewRequest("GET", fmt.Sprintf("%s/status.json", statusPrefix), nil)
			if err != nil {
				return fmt.Errorf("error preparing to fetch status: %v", err)
			}
			am.AddAuthHeader(req)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				return fmt.Errorf("error fetching status: %v", err)
			}
			defer resp.Body.Close()
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return fmt.Errorf("error reading status response: %v", err)
			}
			var st status
			if err := json.Unmarshal(data, &st); err != nil {
				return err
			}
			if st.Version != "" {
				dialogText = fmt.Sprintf("%s\n\nCamlistore %v", dialogText, st.Version)
			}
			if st.GoInfo != "" {
				dialogText = fmt.Sprintf("%s\n\n%v", dialogText, st.GoInfo)
			}
			return nil
		}(); err != nil {
			dom.GetWindow().Alert(fmt.Sprintf("%v", err))
		} else {
			dom.GetWindow().Alert(dialogText)
		}
	}()
}
