package terminfindung

import (
	"DHBW_GO_Projekt/dateisystem"
	"strings"
)

type TerminWithVoted struct {
	Termin       dateisystem.Termin
	VotesFor     int64
	PersonsVoted []string
	isSelected   bool
}

type CorrectForHTML struct {
	TerminID         string
	Info             dateisystem.Termin
	VorschlagTermine []TerminWithVoted
	Persons          []string
	IsLocked         bool
}

func (t TerminFindung) ChangeToCorrectHTML() (rightHTML CorrectForHTML) {
	// create correct HTML
	rightHTML = CorrectForHTML{
		TerminID:         t.Info.ID,
		Info:             t.Info,
		Persons:          []string{},
		VorschlagTermine: []TerminWithVoted{},
		IsLocked:         false,
	}
	//make for to get all the termine
	for _, elem := range t.VorschlagTermine {
		newTermin := TerminWithVoted{
			Termin:     elem,
			isSelected: false,
		}
		//if it is the final termin so it highlights it
		if strings.Compare(elem.ID, t.FinalTermin.ID) == 0 {
			newTermin.isSelected = true
			rightHTML.IsLocked = true
		}
		rightHTML.VorschlagTermine = append(rightHTML.VorschlagTermine, newTermin)
	}
	for person, elem := range t.Persons {
		//append all persons including there names
		rightHTML.Persons = append(rightHTML.Persons, person)
		for id, isVoted := range elem.Votes {
			if isVoted {
				for i, search := range rightHTML.VorschlagTermine {
					if search.Termin.ID == id {
						rightHTML.VorschlagTermine[i].VotesFor += 1
						rightHTML.VorschlagTermine[i].PersonsVoted = append(rightHTML.VorschlagTermine[i].PersonsVoted, person)
					}
				}
			}
		}
	}
	return
}
