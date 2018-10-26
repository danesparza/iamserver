package data_test

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/danesparza/iamserver/data"
)

func TestTOTP_BeginTOTPEnrollment_ValidUser_Successful(t *testing.T) {

	//	Arrange
	systemdb, tokendb := getTestFiles()
	db, err := data.NewManager(systemdb, tokendb)
	if err != nil {
		t.Fatalf("NewManager failed: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(systemdb)
		os.RemoveAll(tokendb)
	}()

	contextUser := data.User{Name: "System"}
	testUser := data.User{Name: "UnitTest1"}
	testPassword := "testpass"

	newUser, err := db.AddUser(contextUser, testUser, testPassword)
	if err != nil {
		t.Errorf("BeginTOTPEnrollment - Should add user without error, but got: %s", err)
	}

	//	Act
	te, err := db.BeginTOTPEnrollment(newUser.Name, 5*time.Minute)

	//	Assert
	if err != nil {
		t.Errorf("BeginTOTPEnrollment - Should begin two-factor enrollment without error, but got: %s", err)
	}

	if te.Secret == "" {
		t.Errorf("BeginTOTPEnrollment failed: Should have a valid TOTP secret: %+v", te)
	}

}

func TestTOTP_BeginTOTPEnrollment_UserDoesntExist_ReturnsError(t *testing.T) {

	//	Arrange
	systemdb, tokendb := getTestFiles()
	db, err := data.NewManager(systemdb, tokendb)
	if err != nil {
		t.Fatalf("NewManager failed: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(systemdb)
		os.RemoveAll(tokendb)
	}()

	//	Act
	_, err = db.BeginTOTPEnrollment("SOME_BOGUS_USER", 5*time.Minute)

	//	Assert
	if err == nil {
		t.Errorf("BeginTOTPEnrollment - Should return error for bogus user, but didn't")
	}

}

func TestTOTP_GetTOTPEnrollment_ValidUser_Successful(t *testing.T) {

	//	Arrange
	systemdb, tokendb := getTestFiles()
	db, err := data.NewManager(systemdb, tokendb)
	if err != nil {
		t.Fatalf("NewManager failed: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(systemdb)
		os.RemoveAll(tokendb)
	}()

	contextUser := data.User{Name: "System"}
	testUser := data.User{Name: "UnitTest1"}
	testPassword := "testpass"

	newUser, err := db.AddUser(contextUser, testUser, testPassword)
	if err != nil {
		t.Errorf("BeginTOTPEnrollment - Should add user without error, but got: %s", err)
	}

	_, err = db.BeginTOTPEnrollment(newUser.Name, 5*time.Minute)
	if err != nil {
		t.Errorf("BeginTOTPEnrollment - Should begin two-factor enrollment without error, but got: %s", err)
	}

	//	Act
	te, err := db.GetTOTPEnrollment(newUser.Name)

	//	Assert
	if err != nil {
		t.Errorf("GetTOTPEnrollmentImage - Should get enrollment without error, but got: %s", err)
	}

	if te.URL == "" {
		t.Errorf("GetTOTPEnrollmentImage - Should get enrollment data, but url is empty")
	}

}

func TestTOTP_GetTOTPEnrollment_ExpiredEnrollment_ReturnsError(t *testing.T) {

	//	Arrange
	systemdb, tokendb := getTestFiles()
	db, err := data.NewManager(systemdb, tokendb)
	if err != nil {
		t.Fatalf("NewManager failed: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(systemdb)
		os.RemoveAll(tokendb)
	}()

	contextUser := data.User{Name: "System"}
	testUser := data.User{Name: "UnitTest1"}
	testPassword := "testpass"

	newUser, err := db.AddUser(contextUser, testUser, testPassword)
	if err != nil {
		t.Errorf("BeginTOTPEnrollment - Should add user without error, but got: %s", err)
	}

	_, err = db.BeginTOTPEnrollment(newUser.Name, 2*time.Second)
	if err != nil {
		t.Errorf("BeginTOTPEnrollment - Should begin two-factor enrollment without error, but got: %s", err)
	}

	//	-- Wait for 5 seconds -- TTL should expire and the token should no longer be available:
	time.Sleep(5 * time.Second)

	//	Act
	_, err = db.GetTOTPEnrollment(newUser.Name)

	//	Assert
	if err == nil {
		t.Errorf("GetTOTPEnrollmentImage - Should return error indicating enrollment not found, but didn't")
	}

	t.Logf("Error from GetTOTPEnrollmentImage: %s", err)

}

func TestTOTP_GetImage_ValidEnrollment_Successful(t *testing.T) {

	//	Arrange
	systemdb, tokendb := getTestFiles()
	db, err := data.NewManager(systemdb, tokendb)
	if err != nil {
		t.Fatalf("NewManager failed: %s", err)
	}
	defer func() {
		db.Close()
		os.RemoveAll(systemdb)
		os.RemoveAll(tokendb)
	}()

	//	-- Clean up the existing 2FA enrollment image
	testEnrollmentImage := "2fa_enrollment.png"
	t.Logf("Removing test enrollment image: %s", testEnrollmentImage)
	os.Remove(testEnrollmentImage)

	contextUser := data.User{Name: "System"}
	testUser := data.User{Name: "UnitTest1"}
	testPassword := "testpass"

	newUser, err := db.AddUser(contextUser, testUser, testPassword)
	if err != nil {
		t.Errorf("BeginTOTPEnrollment - Should add user without error, but got: %s", err)
	}

	_, err = db.BeginTOTPEnrollment(newUser.Name, 5*time.Minute)
	if err != nil {
		t.Errorf("BeginTOTPEnrollment - Should begin two-factor enrollment without error, but got: %s", err)
	}

	te, err := db.GetTOTPEnrollment(newUser.Name)
	if err != nil {
		t.Errorf("GetTOTPEnrollmentImage - Should get enrollment without error, but got: %s", err)
	}

	//	Act
	img, err := te.GetImage()

	//	Assert
	if err != nil {
		t.Errorf("GetImage - Should get enrollment image without error, but got: %s", err)
	}

	if len(img) < 1 {
		t.Errorf("GetTOTPEnrollmentImage - Should get enrollment image, but data is empty")
	}

	t.Logf("Saving test enrollment image: %s", testEnrollmentImage)

	err = ioutil.WriteFile(testEnrollmentImage, img, 0644)
	if err != nil {
		t.Logf("GetImage - Problem saving enrollment image: %s", err)
	}
}
