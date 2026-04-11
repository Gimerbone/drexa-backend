package kycprofile

import "time"

type KycProfile struct {
	documentIdHashed string
	UserId           string
	DateOfBirth      time.Time
	Address          string
	DocumentImage    []byte
	FaceData         []byte
	SubmittedAt      time.Time
	VerifiedAt       time.Time
	Status           string
}

func (k KycProfile) NewKyfProfile(DateOfBirth time.Time, Address string, DocumentImage []byte, FaceData []byte) {

}
