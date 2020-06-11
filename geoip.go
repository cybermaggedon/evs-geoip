//
// GeoIP worker.  Takes events, looks up IP address in GeoIP database, and
// adds location information to the event.  Updated events are transmitted on
// the output queue.
//
// Worker spawns a goroutine which mainly sleeps, and periodically runs
// geoipupdate to update the GeoIP database.
//

package main

import (
	"encoding/binary"
	evs "github.com/cybermaggedon/evs-golang-api"
	"github.com/oschwald/geoip2-golang"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	// How often to update GeoIP data.
	updatePeriod = 86400 * time.Second
)

type GeoIP struct {

	// Embed EventAnalytic framework
	evs.EventAnalytic

	// GeoIP City database
	geoipCityFilename string
	cityDB            *geoip2.Reader

	// GeoIP ASN database
	geoipASNFilename string
	asnDB            *geoip2.Reader

	// Notify channel, notifies analytic that GeoIP databases have been
	// updated and should be reloaded
	notif chan bool
}

// Goroutine: GeoIP updater.  Periodically runs geoipupdate.
func (g *GeoIP) updater() {

	var waitTime = updatePeriod

	for {

		// Wait appropriate sleep period.
		time.Sleep(waitTime)

		log.Print("Running GeoIP update...")

		// Create geoipupdate command.
		cmd := exec.Command("geoipupdate", "-f", "GeoIP.conf",
			"-d", ".")

		// Execute, stdout/stderr to byte array.
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Print("Update error: %s", err.Error())
			log.Print("geoipupdate: %s", out)

			// Failed: Retry sooner than the long period.
			waitTime = 600 * time.Second
			continue

		}

		log.Print("GeoIP updated, success.")

		// On successful update, wait period is a long period.
		waitTime = updatePeriod

		// Ping the main goroutine, so it knows to reopen the
		// GeoIP database.
		g.notif <- true

	}

}

// Open GeoIP databases.
func (s *GeoIP) openGeoIP() {

	log.Print("Opening GeoIP databases")

	// Open database.
	cityDB, err := geoip2.Open(s.geoipCityFilename)

	// If ok...
	if err == nil {
		// ...store database handle and return.
		s.cityDB = cityDB
	}

	// Open database.
	asnDB, err := geoip2.Open(s.geoipASNFilename)

	// If ok...
	if err == nil {
		// ...store database handle and return.
		s.asnDB = asnDB
	}

	if s.cityDB == nil {
		log.Print("No active GeoIP city DB")
	} else {
		log.Print("GeoIP City is enabled")
	}

	if s.asnDB == nil {
		log.Print("No active GeoIP ASN DB")
	} else {
		log.Print("GeoIP ASN is enabled")
	}

}

// Initialisation
func (g *GeoIP) Init(binding string, outputs []string) error {

	g.notif = make(chan bool)

	// Database filenames are environment variables.
	var ok bool
	if g.geoipCityFilename, ok = os.LookupEnv("GEOIP_DB"); !ok {
		g.geoipCityFilename = "GeoLite2-City.mmdb"
	}
	if g.geoipASNFilename, ok = os.LookupEnv("GEOIP_ASN_DB"); !ok {
		g.geoipASNFilename = "GeoLite2-ASN.mmdb"
	}

	// Open databases.
	g.openGeoIP()

	g.EventAnalytic.Init(binding, outputs, g)

	return nil

}

// GeoIP lookup
func (g *GeoIP) lookup(ip net.IP) (*evs.Locations_Location, error) {

	// Lookup in GeoIP database.
	var city *geoip2.City
	var asn *geoip2.ASN
	var err error
	if g.cityDB != nil {
		city, err = g.cityDB.City(ip)
		if err != nil {
			return nil, err
		}
	}

	if g.asnDB != nil {
		// Lookup in ASN database
		asn, err = g.asnDB.ASN(ip)
		if err != nil {
			return nil, err
		}
	}

	// If nil return, give up.
	if city == nil && asn == nil {
		return nil, nil
	}

	// Get data from GeoIP record.
	locn := &evs.Locations_Location{}

	if city != nil {
		locn.City = city.City.Names["en"]
		locn.Iso = city.Country.IsoCode
		locn.Country = city.Country.Names["en"]
		locn.Latitude = float32(city.Location.Latitude)
		locn.Longitude = float32(city.Location.Longitude)
		//	locn.AccuracyRadius = int(city.Location.AccuracyRadius)
		locn.Postcode = city.Postal.Code
	}

	if asn != nil {
		locn.Asnum = strconv.Itoa(int(asn.AutonomousSystemNumber))
		locn.Asorg = asn.AutonomousSystemOrganization
	}

	// Don't return an empty record.
	if locn.City == "" && locn.Iso == "" && locn.Country == "" &&
		locn.Latitude == 0.0 &&
		locn.Longitude == 0.0 &&
		locn.Postcode == "" {
		return nil, nil
	}

	// Return the complete record.
	return locn, nil

}

// Converts a 32-bit int to an IP address
// FIXME: Copied from detector, put in a library
func int32ToIp(ipLong uint32) net.IP {
	ipByte := make([]byte, 4)
	binary.BigEndian.PutUint32(ipByte, ipLong)
	return net.IP(ipByte)
}

// Converts a byte array to an IP address. This is for IPv6 addresses.
func bytesToIp(b []byte) net.IP {
	return net.IP(b)
}

// Event handler for new events.
func (g *GeoIP) Event(ev *evs.Event, properties map[string]string) error {

	log.Print("Event")

	// If there's a signal from the GeoIP database updater, re-open the
	// database.
	select {
	case _ = <-g.notif:
		log.Print("Update occured - reopening GeoIP database")
		g.openGeoIP()

	default:
		// No signal, do nothing.
	}

	var src, dest net.IP

	for _, addr := range ev.Src {
		if addr.Protocol == evs.Protocol_ipv4 {
			src = int32ToIp(addr.Address.GetIpv4())
			break
		}
		if addr.Protocol == evs.Protocol_ipv6 {
			src = bytesToIp(addr.Address.GetIpv6())
			break
		}
	}

	for _, addr := range ev.Dest {
		if addr.Protocol == evs.Protocol_ipv4 {
			dest = int32ToIp(addr.Address.GetIpv4())
			break
		}
		if addr.Protocol == evs.Protocol_ipv6 {
			dest = bytesToIp(addr.Address.GetIpv6())
			break
		}
	}

	// Get location information from IP addresses.
	srcloc, _ := g.lookup(src)
	destloc, _ := g.lookup(dest)

	// If we get either a source or destination location, store the
	// information in the event record.
	if srcloc != nil || destloc != nil {
		ev.Location = &evs.Locations{}
		if srcloc != nil {
			ev.Location.Src = srcloc
		}
		if destloc != nil {
			ev.Location.Dest = destloc
		}
	}

	g.OutputEvent(ev, properties)


	return nil

}

func main() {

	g := &GeoIP{}
	// Notification channel.  A bool gets sent down the channel every time
	// the updater goroutine inovkes an update.
	g.notif = make(chan bool, 2)

	// Launch updater goroutine
	go g.updater()

	binding, ok := os.LookupEnv("INPUT")
	if !ok {
		binding = "cyberprobe"
	}

	out, ok := os.LookupEnv("OUTPUT")
	if !ok {
		g.Init(binding, []string{"geo"})
	} else {
		outarray := strings.Split(out, ",")
		g.Init(binding, outarray)
	}

	log.Print("Initialisation complete")

	g.Run()

}
