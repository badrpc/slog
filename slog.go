// Package slog provides alternative syslog client API. An internal
// syslog writer used to send messages to a syslog service with options
// to tune it.
//
// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package slog

import (
	"fmt"
	"log"
	"log/syslog"
	"strings"
	"sync/atomic"
	"unsafe"
)

var (
	unsafeSyslogWriter unsafe.Pointer // Always *syslog.Writer

	noInitWarningDone       bool
	failedSyslogWarningDone bool
)

type params struct {
	network  string
	raddr    string
	facility syslog.Priority
	tag      string
}

type Option func(p *params)

const facilityStrPrefix = "LOG_"

// ParseFacility converts string representation of a syslog facility into
// syslog.Priority value. The standard facilities as described by FreeBSD
// `man syslog' as of 12.0-RELEASE are recognised (LOG_DAEMON, LOG_USER, etc).
// Parsing is case insensitive and LOG_ prefix is optional and can be omitted.
func ParseFacility(facility string) (syslog.Priority, error) {
	f := strings.ToUpper(facility)
	if strings.HasPrefix(f, facilityStrPrefix) {
		f = f[len(facilityStrPrefix):]
	}
	switch f {
	case "KERN":
		return syslog.LOG_KERN, nil
	case "USER":
		return syslog.LOG_USER, nil
	case "MAIL":
		return syslog.LOG_MAIL, nil
	case "DAEMON":
		return syslog.LOG_DAEMON, nil
	case "AUTH":
		return syslog.LOG_AUTH, nil
	case "SYSLOG":
		return syslog.LOG_SYSLOG, nil
	case "LPR":
		return syslog.LOG_LPR, nil
	case "NEWS":
		return syslog.LOG_NEWS, nil
	case "UUCP":
		return syslog.LOG_UUCP, nil
	case "CRON":
		return syslog.LOG_CRON, nil
	case "AUTHPRIV":
		return syslog.LOG_AUTHPRIV, nil
	case "FTP":
		return syslog.LOG_FTP, nil
	case "LOCAL0":
		return syslog.LOG_LOCAL0, nil
	case "LOCAL1":
		return syslog.LOG_LOCAL1, nil
	case "LOCAL2":
		return syslog.LOG_LOCAL2, nil
	case "LOCAL3":
		return syslog.LOG_LOCAL3, nil
	case "LOCAL4":
		return syslog.LOG_LOCAL4, nil
	case "LOCAL5":
		return syslog.LOG_LOCAL5, nil
	case "LOCAL6":
		return syslog.LOG_LOCAL6, nil
	case "LOCAL7":
		return syslog.LOG_LOCAL7, nil
	}
	return 0, fmt.Errorf("cannot parse %q as syslog facility", facility)
}

// WithFacility is an option for Init which adjusts facility in outgoing syslog
// messages.
func WithFacility(facility syslog.Priority) Option {
	return func(p *params) {
		p.facility = facility
	}
}

// WithTag is an option for Init which adjusts tag in outgoing syslog messages.
// Tag is passed directly to syslog.Dial, see corresponding documentation for
// more details.
func WithTag(tag string) Option {
	return func(p *params) {
		p.tag = tag
	}
}

// WithTag is an option for Init to specify parameter syslog service connection.
// Both values will be passed to syslog.Dia, see corresponding documentation for
// more details. As of the time this text being written, an empty value of
// network parameter requests a connection over a UNIX socket to a local syslog
// service (raddr is ignored in this case). Alternatively network can be a
// string accepted by net.Dial.
func WithDial(network, raddr string) Option {
	return func(p *params) {
		p.network = network
		p.raddr = raddr
	}
}

// Init initializes or re-initializes internal syslog writer. It is expected
// to be safe to call this function from concurrent goroutines.
func Init(opts ...Option) error {
	var p params
	for _, o := range opts {
		o(&p)
	}

	var w *syslog.Writer
	var err error
	if p.network == "" {
		w, err = syslog.New(p.facility, p.tag)
	} else {
		w, err = syslog.Dial(p.network, p.raddr, p.facility, p.tag)
	}
	if err == nil {
		old := (*syslog.Writer)(atomic.SwapPointer(&unsafeSyslogWriter, unsafe.Pointer(w)))
		if old != nil {
			old.Close()
		}
	}
	return err
}

// Alert sends a syslog message with severity LOG_ALERT.
func Alert(v ...interface{}) {
	write(fmt.Sprint(v...), (*syslog.Writer).Alert)
}

// Alertf sends a formatted syslog message with severity LOG_ALERT.
func Alertf(format string, v ...interface{}) {
	write(fmt.Sprintf(format, v...), (*syslog.Writer).Alert)
}

// Crit sends a syslog message with severity LOG_CRIT.
func Crit(v ...interface{}) {
	write(fmt.Sprint(v...), (*syslog.Writer).Crit)
}

// Critf sends a formatted syslog message with severity LOG_CRIT.
func Critf(format string, v ...interface{}) {
	write(fmt.Sprintf(format, v...), (*syslog.Writer).Crit)
}

// Debug sends a syslog message with severity LOG_DEBUG.
func Debug(v ...interface{}) {
	write(fmt.Sprint(v...), (*syslog.Writer).Debug)
}

// Debugf sends a formatted syslog message with severity LOG_DEBUG.
func Debugf(format string, v ...interface{}) {
	write(fmt.Sprintf(format, v...), (*syslog.Writer).Debug)
}

// Emerg sends a syslog message with severity LOG_EMERG.
func Emerg(v ...interface{}) {
	write(fmt.Sprint(v...), (*syslog.Writer).Emerg)
}

// Emergf sends a formatted syslog message with severity LOG_EMERG.
func Emergf(format string, v ...interface{}) {
	write(fmt.Sprintf(format, v...), (*syslog.Writer).Emerg)
}

// Err sends a syslog message with severity LOG_ERR.
func Err(v ...interface{}) {
	write(fmt.Sprint(v...), (*syslog.Writer).Err)
}

// Errf sends a formatted syslog message with severity LOG_ERR.
func Errf(format string, v ...interface{}) {
	write(fmt.Sprintf(format, v...), (*syslog.Writer).Err)
}

// Info sends a syslog message with severity LOG_INFO.
func Info(v ...interface{}) {
	write(fmt.Sprint(v...), (*syslog.Writer).Info)
}

// Infof sends a formatted syslog message with severity LOG_INFO.
func Infof(format string, v ...interface{}) {
	write(fmt.Sprintf(format, v...), (*syslog.Writer).Info)
}

// Notice sends a syslog message with severity LOG_NOTICE.
func Notice(v ...interface{}) {
	write(fmt.Sprint(v...), (*syslog.Writer).Notice)
}

// Noticef sends a formatted syslog message with severity LOG_NOTICE.
func Noticef(format string, v ...interface{}) {
	write(fmt.Sprintf(format, v...), (*syslog.Writer).Notice)
}

// Warning sends a syslog message with severity LOG_WARNING.
func Warning(v ...interface{}) {
	write(fmt.Sprint(v...), (*syslog.Writer).Warning)
}

// Warningf sends a formatted syslog message with severity LOG_WARNING.
func Warningf(format string, v ...interface{}) {
	write(fmt.Sprintf(format, v...), (*syslog.Writer).Warning)
}

func syslogWriter() *syslog.Writer {
	return (*syslog.Writer)(atomic.LoadPointer(&unsafeSyslogWriter))
}

func write(message string, method func(*syslog.Writer, string) error) {
	sw := syslogWriter()
	if sw == nil {
		if !noInitWarningDone {
			log.Print("Log requests before syslog.Init are sent to default log.")
			noInitWarningDone = true
		}
		log.Print(message)
		return
	}
	if err := method(sw, message); err != nil {
		if !failedSyslogWarningDone {
			log.Print("Error sending message to syslog: ", err)
			failedSyslogWarningDone = true
		}
		log.Print(message)
		return
	}
	failedSyslogWarningDone = false
}
