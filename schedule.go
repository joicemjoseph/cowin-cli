package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/olekukonko/tablewriter"
)

type CenterBookable struct {
	Name        string
	Freetype    string
	SessionID   string
	MinAgeLimit int
	Date        string
	Vaccine     string
}

type beneficariesData struct {
	Beneficiaries []struct {
		BeneficiaryReferenceID string `json:"beneficiary_reference_id"`
		Name                   string `json:"name"`
		Dose1Date              string `json:"dose1_date"`
	} `json:"beneficiaries"`
}

type ScheduleData struct {
	txnId              string
	bearerToken        string
	beneficariesRefIDs []string
	dose               int
	sessionID          string
}

func checkError(err error) {
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func getBeneficaries(bearerToken string) beneficariesData {
	var b beneficariesData
	resp, statusCode := getReqAuth(beneficiariesURL, bearerToken)

	if statusCode != 200 {
		log.Fatalln("Cannot get beneficaries")
	}

	json.Unmarshal(resp, &b)

	return b

}

func getUserSelection(message string, limit int, all bool) int {
	var opt int
	again := false
	fmt.Println()
	for {
		if again {
			fmt.Println("Wrong selection")
		}
		fmt.Print(message)
		fmt.Scanf("%d\n", &opt)

		if opt <= limit || (all && opt == limit+1) {
			break
		} else {
			again = true
		}
	}
	fmt.Println()
	return opt
}

func getDoseNo(doseDate string) int {
	if doseDate == "" {
		return 1
	}
	return 2
}

func (scheduleData *ScheduleData) getAllbID(b beneficariesData) {
	for _, v := range b.Beneficiaries {
		scheduleData.beneficariesRefIDs = append(scheduleData.beneficariesRefIDs, v.BeneficiaryReferenceID)
	}
	scheduleData.dose = getDoseNo(b.Beneficiaries[0].Dose1Date)
}

func printBeneficaries(b beneficariesData) {
	var all int
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Name"})

	for i, v := range b.Beneficiaries {
		table.Append([]string{fmt.Sprint(i), v.Name})
		all = i
	}
	table.Append([]string{fmt.Sprint(all + 1), "All"})

	table.Render()

}

// getBeneficariesID gets list of beneficaries id and dose
func (scheduleData *ScheduleData) getBeneficariesID(b beneficariesData, name string) {
	var limit, opt int

	if len(b.Beneficiaries) == 1 {
		scheduleData.beneficariesRefIDs = append(scheduleData.beneficariesRefIDs, b.Beneficiaries[0].BeneficiaryReferenceID)
		scheduleData.dose = getDoseNo(b.Beneficiaries[0].Dose1Date)
		// name specified
	} else if name != "" {
		// get all beneficaries
		if name == "all" {
			scheduleData.getAllbID(b)
		} else {
			for _, v := range b.Beneficiaries {
				if v.Name == name {
					scheduleData.beneficariesRefIDs = append(scheduleData.beneficariesRefIDs, v.BeneficiaryReferenceID)
					scheduleData.dose = getDoseNo(v.Dose1Date)
					break
				}

			}
			if len(scheduleData.beneficariesRefIDs) == 0 {
				log.Fatalf("name %v not found\n", name)
			}
		}

	} else {
		//print beneficaries and prompt user
		printBeneficaries(b)
		opt = getUserSelection("Enter name ID : ", limit, true)

		// get all beneficaries
		if opt == limit+1 {
			scheduleData.getAllbID(b)
			// append chosen one
		} else {
			scheduleData.beneficariesRefIDs = append(scheduleData.beneficariesRefIDs, b.Beneficiaries[opt].BeneficiaryReferenceID)
			scheduleData.dose = getDoseNo(b.Beneficiaries[opt].Dose1Date)
		}
	}

}

// printCenterBookable prints centers avaliable for booking
func printCenterBookable(centerList []CenterBookable) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ID", "Center", "Free type", "Min Age", "Date"})
	for i, v := range centerList {
		table.Append([]string{fmt.Sprint(i), v.Name, v.Freetype, fmt.Sprint(v.MinAgeLimit), v.Date})
	}
	table.Render()
}

func (scheduleData *ScheduleData) getSessionID(districtID, date string) {
	var center CentreData
	var centerBookable []CenterBookable
	var opt int
	center.getCenters(districtID, "", "", date)
	// Try again if centers is empty
	if len(center.Centers) < 1 {
		center.getCenters(districtID, "", "", date)
	}

	for _, v := range center.Centers {
		for _, vv := range v.Sessions {
			if vv.AvailableCapacity > 0 {
				centerBookable = append(centerBookable, CenterBookable{
					Name:        v.Name,
					Freetype:    v.FeeType,
					SessionID:   vv.SessionID,
					Vaccine:     vv.Vaccine,
					MinAgeLimit: vv.MinAgeLimit,
					Date:        vv.Date,
				})

			}
		}
	}
	if len(centerBookable) > 0 {
		printCenterBookable(centerBookable)
		opt = getUserSelection("Enter Center ID :", len(centerBookable), false)

		scheduleData.sessionID = centerBookable[opt].SessionID
	} else {
		log.Fatalln("No Centers available for booking")
	}

}

func (scheduleData ScheduleData) scheduleVaccineNow() {
	postData := map[string]interface{}{
		"dose":          scheduleData.dose,
		"session_id":    scheduleData.sessionID,
		"slot":          "FORENOON",
		"beneficiaries": scheduleData.beneficariesRefIDs,
	}
	jsonBytes, _ := json.Marshal(postData)

	_, statusCode := postReq(appointmentSchedule, jsonBytes, scheduleData.bearerToken)

	switch statusCode {
	case 200:
		fmt.Println("Appointment scheduled successfully")
	case 400:
		fmt.Println("Bad Request")
	case 401:
		fmt.Println("Unauthenticated Access")
	case 409:
		fmt.Println("This vaccination center is completely booked for the selected date")
	case 500:
		fmt.Println("Internal Server error")
	default:
		log.Fatalln("Error")
	}

}

func scheduleVaccine(districtID, pincode, date, mobileNumber, name string) {
	var scheduleData ScheduleData

	scheduleData.genOTP(mobileNumber)

	scheduleData.getSessionID(districtID, date)

	scheduleData.validateOTP(getOTPprompt())

	if scheduleData.bearerToken == "" {
		fmt.Println("Incorrect OTP")
		scheduleData.validateOTP(getOTPprompt())
	}

	scheduleData.getBeneficariesID(getBeneficaries(scheduleData.bearerToken), name)

	scheduleData.scheduleVaccineNow()

}
