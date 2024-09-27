package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"

	webview "github.com/webview/webview_go"
)

type Vehicle struct {
	Name             string  // vehicle identifier (14 characters max) (A)
	Mass             float64 // fMass (B)
	Drag             float64 // fDragMult (C)
	PercentSubmerged float64 // nPercentSubmerged (D) [10 to 120]
	ComX             float64 // CentreOfMass.x (E) [-10.0 > x > 10.0]
	ComY             float64 // CentreOfMass.y (F) [-10.0 > x > 10.0]
	ComZ             float64 // CentreOfMass.z (G) [-10.0 > x > 10.0]

	// Transmission
	DriveBias      float64 // m_nDriveBias (Tt) [1.0 = FWD] [0.0 = RWD]
	DriveGears     int     // m_nDriveGears (Tg)
	DriveForce     float64 // m_fDriveForce (Tf)
	DriveInertia   float64 // m_fDriveInertia (Ti)
	MaxVelocity    float64 // m_fV (Tv)
	BrakeForce     float64 // m_fBrakeForce (Tb)
	BrakeBias      float64 // m_fBrakeBias (Tbb)
	BrakeSomething float64 // undocumented (Thb)
	SteeringLock   float64 // m_fSteeringLock (Ts)

	// Wheel Traction
	TractionCurveMax     float64 // m_fTractionCurveMax (Wc+)
	TractionCurveMin     float64 // m_fTractionCurveMin (Wc-)
	TractionCurveLateral float64 // m_fTractionCurveLateral (Wc-)
	///////////////////////////////// m_fTractionCurveLongitudinal (Wc|) (shape of longituduinal traction curve (peak traction position in degrees)
	TractionSpringDeltaMax float64 // m_fTractionSpringDeltaMax (Ws+) (max dist for traction spring)
	TractionBias           float64 // m_fTractionBias (Wh)

	// Suspension
	SuspensionForce       float64 // m_fSuspensionForce (Sf) (1 / (Force * NumWheels) = Lower limit for zero force at full extension
	SuspensionCompDamp    float64 // m_fSuspensionCompDamp (Scd)
	SuspensionReboundDamp float64 // m_fSuspensionReboundDamp (Srd)
	SuspensionUpperLimit  float64 // m_fSuspensionUpperLimit (Su)
	SuspensionLowerLimit  float64 // m_fSuspensionLowerLimit (Sl)
	SuspensionRaise       float64 // m_fSuspensionRaise (Sr)
	SuspensionBias        float64 // m_fSuspensionBias (Sb)

	// Damage
	CollisionDamageMult   float64 // m_fCollisionDamageMult (Dc)
	WeaponDamageMult      float64 // m_fWeaponDamageMult (Dw)
	DeformationDamageMult float64 // m_fDeformationDamageMult (Dd)
	EngineDamageMult      float64 // m_fEngineDamageMult (De)

	// Misc
	SeatOffsetDist float64 // m_fSeatOffsetDist (Ms)
	MonetaryValue  int     // m_nMonetaryValue (Mv)
	ModelFlags     string  // mFlags (Mmf) MODEL FLAGS in hex
	HandlingFlags  string  // hFlags (Mhf) HANDLING FLAGS in hex
	AnimGroup      string  // m_nAnimGroup (Ma) anim group type
}

var current string       // current handling.dat
var specialJunk []string // boat handling, bike handling, plane handling, etc. TODO implement?

func main() {
	w := webview.New(true)
	defer w.Destroy()

	w.SetTitle("guh editor")
	w.SetSize(600, 400, webview.HintNone)
	w.Bind("processFile", func(content string) {
		current = content
		fmt.Println("file loaded")
	})

	w.Bind("applyBomb", func(offset_in string, options map[string]bool) {
		offset, _ := strconv.ParseFloat(offset_in, 64)
		offset /= 100

		if current == "" {
			fmt.Println("no file loaded")
			return
		}

		vehicles := parseContent(current)
		log.Println("parseContent() ok")
		fmt.Println(vehicles[0])

		rainbomb(vehicles, 0, offset, options) // Pass options
		log.Println("rainbomb() ok")

		writeFile(vehicles, "ohandling.dat")
		log.Println("writeFile() ok")

		fmt.Println("ohandling.dat written")
	})

	htmlContent, err := os.ReadFile("index.html")
	if err != nil {
		fmt.Println("No index.html:", err)
		return
	}
	w.SetHtml(string(htmlContent))
	//w.Navigate(`data:text/html,` + string(htmlContent))
	w.Run()
}

func rainbomb(vehicles []Vehicle, minOffset, maxOffset float64, settings map[string]bool) {
	for i := range vehicles {
		// Apply only if the corresponding checkbox is checked
		if settings["driveforce"] {
			applyDriveForce(&vehicles[i], minOffset, maxOffset)
		}
		if settings["brakeforce"] {
			applyBrakeForce(&vehicles[i], minOffset, maxOffset)
		}
		if settings["traction"] {
			applyTraction(&vehicles[i], minOffset, maxOffset)
		}
		if settings["maxVelocity"] {
			applyMaxVelocity(&vehicles[i], minOffset, maxOffset)
		}
		if settings["drag"] {
			applyDrag(&vehicles[i], minOffset, maxOffset)
		}
	}
}

func applyDriveForce(v *Vehicle, minOffset, maxOffset float64) {
	v.DriveForce = getOffsetValue(v.DriveForce, minOffset, maxOffset)
}

func applyBrakeForce(v *Vehicle, minOffset, maxOffset float64) {
	v.BrakeForce = getOffsetValue(v.BrakeForce, minOffset, maxOffset)
}

// traction curve min (Wc-)
func applyTraction(v *Vehicle, minOffset, maxOffset float64) {
	v.TractionCurveMin = getOffsetValue(v.TractionCurveMin, minOffset, maxOffset)
}

func applyMaxVelocity(v *Vehicle, minOffset, maxOffset float64) {
	v.MaxVelocity = getOffsetValue(v.MaxVelocity, minOffset, maxOffset)
}

func applyDrag(v *Vehicle, minOffset, maxOffset float64) {
	offset := minOffset + rand.Float64()*(maxOffset-minOffset)
	v.Drag = v.Drag / (1 + offset)
}

/*
func applyOffset(v *Vehicle, minOffset, maxOffset float64) {
	val := reflect.ValueOf(v).Elem()

	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		if field.Kind() == reflect.Float64 {
			original := field.Float()
			newValue := getOffsetValue(original, minOffset, maxOffset)
			field.SetFloat(newValue)
		} else if field.Kind() == reflect.Int {
			original := float64(field.Int())
			newValue := getOffsetValue(original, minOffset, maxOffset)
			field.SetInt(int64(newValue))
		}
	}
}
*/

func getOffsetValue(value, minOffset, maxOffset float64) float64 {
	offset := minOffset + rand.Float64()*(maxOffset-minOffset)
	return value * (1 + offset)
}

func writeFile(vehicles []Vehicle, filePath string) {
	file, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)

	for _, v := range vehicles {
		//fmt.Println(v)

		val := reflect.ValueOf(v)
		var fields []string
		for i := 0; i < val.NumField(); i++ {
			field := val.Field(i)

			switch field.Kind() {
			case reflect.String:
				fields = append(fields, field.String())
			case reflect.Int:
				fields = append(fields, fmt.Sprintf("%d", field.Int()))
			case reflect.Float64:
				fields = append(fields, fmt.Sprintf("%.2f", field.Float()))
			default:
				log.Printf("Unsupported field: %d: %s", i, field.Kind())
			}
		}
		//fields = append(fields, fmt.Sprintf("%s", v.AnimGroup))
		line := strings.Join(fields, " ") + "\n"
		writer.WriteString(line)
	}
	for _, line := range specialJunk {
		writer.WriteString(line + "\n")
	}
	writer.Flush()
}

func parseContent(content string) []Vehicle {
	vehicles := []Vehicle{}
	//scanner := bufio.NewScanner(file)
	scanner := bufio.NewScanner(strings.NewReader(content))

	for scanner.Scan() {
		//line := scanner.Text()
		line := strings.TrimSpace(scanner.Text())
		//fmt.Println(line)
		//if len(strings.TrimSpace(line)) == 0 || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
		if len(line) == 0 || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			// comment/empty line
			//writer.WriteString(line + "\n")
			continue
		}

		if strings.HasPrefix(line, "%") || strings.HasPrefix(line, "!") ||
			strings.HasPrefix(line, "$") || strings.HasPrefix(line, "^") {
			specialJunk = append(specialJunk, line)
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 36 {
			continue // skip lines that dont have enough fields ???
		}

		mass, _ := strconv.ParseFloat(fields[1], 64)          // 1
		drag, _ := strconv.ParseFloat(fields[2], 64)          // 2
		submerged, _ := strconv.ParseFloat(fields[3], 64)     // 3
		comX, _ := strconv.ParseFloat(fields[4], 64)          // 4
		comY, _ := strconv.ParseFloat(fields[5], 64)          // 5
		comZ, _ := strconv.ParseFloat(fields[6], 64)          // 6
		driveBias, _ := strconv.ParseFloat(fields[7], 64)     // 7
		driveGears, _ := strconv.Atoi(fields[8])              // 8 OK
		driveForce, _ := strconv.ParseFloat(fields[9], 64)    // 9 OK
		driveInertia, _ := strconv.ParseFloat(fields[10], 64) // 10 OK
		maxVelocity, _ := strconv.ParseFloat(fields[11], 64)  // 11 OK
		brakeForce, _ := strconv.ParseFloat(fields[12], 64)   // 12 OK
		brakeBias, _ := strconv.ParseFloat(fields[13], 64)    // 13 OK
		thb, _ := strconv.ParseFloat(fields[14], 64)          // 14 NEW: Thb value (Missing before)

		steeringLock, _ := strconv.ParseFloat(fields[15], 64) // 15 OK

		tractionCurveMax, _ := strconv.ParseFloat(fields[16], 64)       // 16 OK
		tractionCurveMin, _ := strconv.ParseFloat(fields[17], 64)       // 17 OK
		tractionCurveLateral, _ := strconv.ParseFloat(fields[18], 64)   // 18 OK
		tractionSpringDeltaMax, _ := strconv.ParseFloat(fields[19], 64) // 19 OK
		tractionBias, _ := strconv.ParseFloat(fields[20], 64)           // 20 OK

		suspensionForce, _ := strconv.ParseFloat(fields[21], 64)       // 21 OK
		suspensionCompDamp, _ := strconv.ParseFloat(fields[22], 64)    // 22 OK
		suspensionReboundDamp, _ := strconv.ParseFloat(fields[23], 64) // 23 OK
		suspensionUpperLimit, _ := strconv.ParseFloat(fields[24], 64)  // 24 OK
		suspensionLowerLimit, _ := strconv.ParseFloat(fields[25], 64)  // 25 OK
		suspensionRaise, _ := strconv.ParseFloat(fields[26], 64)       // 26 OK
		suspensionBias, _ := strconv.ParseFloat(fields[27], 64)        // 27 OK

		collisionDamageMult, _ := strconv.ParseFloat(fields[28], 64)   // 28 OK
		weaponDamageMult, _ := strconv.ParseFloat(fields[29], 64)      // 29 OK
		deformationDamageMult, _ := strconv.ParseFloat(fields[30], 64) // 30 OK
		engineDamageMult, _ := strconv.ParseFloat(fields[31], 64)      // 31 OK

		seatOffsetDist, _ := strconv.ParseFloat(fields[32], 64) // 32 OK
		monetaryValue, _ := strconv.Atoi(fields[33])            // 33 OK

		//vehicles = append(vehicles, Vehicle{
		vehicle := Vehicle{
			Name:                   fields[0],
			Mass:                   mass,
			Drag:                   drag,
			PercentSubmerged:       submerged,
			ComX:                   comX,
			ComY:                   comY,
			ComZ:                   comZ,
			DriveBias:              driveBias,
			DriveGears:             driveGears,
			DriveForce:             driveForce,
			DriveInertia:           driveInertia,
			MaxVelocity:            maxVelocity,
			BrakeForce:             brakeForce,
			BrakeBias:              brakeBias,
			BrakeSomething:         thb,
			SteeringLock:           steeringLock,
			TractionCurveMax:       tractionCurveMax,
			TractionCurveMin:       tractionCurveMin,
			TractionCurveLateral:   tractionCurveLateral,
			TractionSpringDeltaMax: tractionSpringDeltaMax,
			TractionBias:           tractionBias,
			SuspensionForce:        suspensionForce,
			SuspensionCompDamp:     suspensionCompDamp,
			SuspensionReboundDamp:  suspensionReboundDamp,
			SuspensionUpperLimit:   suspensionUpperLimit,
			SuspensionLowerLimit:   suspensionLowerLimit,
			SuspensionRaise:        suspensionRaise,
			SuspensionBias:         suspensionBias,
			CollisionDamageMult:    collisionDamageMult,
			WeaponDamageMult:       weaponDamageMult,
			DeformationDamageMult:  deformationDamageMult,
			EngineDamageMult:       engineDamageMult,
			SeatOffsetDist:         seatOffsetDist,
			MonetaryValue:          monetaryValue,
			ModelFlags:             fields[34],
			HandlingFlags:          fields[35],
			AnimGroup:              fields[36],
		}

		//log.Printf("Parsed vehicle: %+v", vehicle) // Debugging line
		vehicles = append(vehicles, vehicle)

		if err := scanner.Err(); err != nil {
			fmt.Println("bufio.scanner error:", err)
		}
	}
	return vehicles
}
