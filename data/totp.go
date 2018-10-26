package data

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"time"

	"github.com/danesparza/badger"
	"github.com/danesparza/otp"
	"github.com/danesparza/otp/totp"
)

// TotpEnrollment represents an enrollment record for
// Time based one-time-pad (two factor authentication) for a user
// Both the secret and the image will be stored temporarily until
// the user validates the key with a generated password (indicating
// they have setup the TOTP key in their app and have generated a valid
// code at least once).  When enrollment is complete, this record will be
// removed and the secret will be stored with the user data
type TotpEnrollment struct {
	User   string `json:"user"`
	Secret string `json:"secret"`
	Image  string `json:"image"`
	URL    string `json:"url"`
}

// BeginTOTPEnrollment begins TOTP enrollment for a user.  If the user already has two factor
// authentication enabled, this will return an error
func (store Manager) BeginTOTPEnrollment(userName string, expiresafter time.Duration) (TotpEnrollment, error) {
	//	Our return item
	retval := TotpEnrollment{}

	//	The user to check
	user := User{}

	//	First -- find out if the user is already enrolled in two-factor authentication
	err := store.systemdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(GetKey("User", userName))
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}

		if len(val) > 0 {
			//	Unmarshal data into our item
			if err := json.Unmarshal(val, &user); err != nil {
				return err
			}
		}

		return nil
	})

	//	If we got an error, we have a problem:
	if err != nil {
		return retval, fmt.Errorf("User does not exist")
	}

	//	If the user is already enrolled -- return an error
	if user.TOTPEnabled == true {
		return retval, fmt.Errorf("User already has TOTP enabled.  To get a new TOTP key, disable TOTP first -- then re-enroll")
	}

	//	Generate the TOTP information:
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "IAMServer",
		AccountName: userName,
	})
	if err != nil {
		return retval, fmt.Errorf("Problem generating TOTP key: %s", err)
	}

	//	Get the image for the TOTP enrollment
	var buf bytes.Buffer
	img, err := key.Image(200, 200)
	if err != nil {
		panic(err)
	}
	png.Encode(&buf, img)

	//	Convert the image data to a base64 string
	encodedImage := base64.StdEncoding.EncodeToString(buf.Bytes())

	//	Store the secret, url, and image
	retval = TotpEnrollment{
		User:   userName,
		Secret: key.Secret(),
		Image:  encodedImage,
		URL:    key.URL(),
	}

	//	Serialize to JSON format
	encoded, err := json.Marshal(retval)
	if err != nil {
		return retval, fmt.Errorf("Problem serializing the data: %s", err)
	}

	//	Save it to the database:
	err = store.systemdb.Update(func(txn *badger.Txn) error {
		err := txn.SetWithTTL(GetKey("TotpEnrollment", retval.User), encoded, expiresafter)
		return err
	})

	//	Return our data:
	return retval, nil
}

// FinishTOTPEnrollment finishes TOTP enrollment for a user.  If the user already has two factor
// authentication enabled, this will return an error
func (store Manager) FinishTOTPEnrollment(userName, validationCode string) (User, error) {
	//	Our return item
	enrollment := TotpEnrollment{}

	//	The user to check
	user := User{}

	//	First, make sure we can look up the user's enrollment:
	err := store.systemdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(GetKey("TotpEnrollment", userName))
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}

		if len(val) > 0 {
			//	Unmarshal data into our item
			if err := json.Unmarshal(val, &enrollment); err != nil {
				return err
			}
		}

		return nil
	})

	//	If we got an error, we have a problem:
	if err != nil {
		return user, fmt.Errorf("Enrollment not found")
	}

	//	Next -- find out if the user is already enrolled in two-factor authentication
	err = store.systemdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(GetKey("User", userName))
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}

		if len(val) > 0 {
			//	Unmarshal data into our item
			if err := json.Unmarshal(val, &user); err != nil {
				return err
			}
		}

		return nil
	})

	//	If we got an error, we have a problem:
	if err != nil {
		return user, fmt.Errorf("User does not exist")
	}

	//	If the user is already enrolled -- return an error
	if user.TOTPEnabled == true {
		return user, fmt.Errorf("User already has TOTP enabled.  To get a new TOTP key, disable TOTP and then re-enroll")
	}

	//	Validate the TOTP information:
	validEnrollment := totp.Validate(validationCode, enrollment.Secret)
	if !validEnrollment {
		return user, fmt.Errorf("Not a valid OTP code.  Please use the code from your authentication app")
	}

	//	Set the secret and turn on two factor for the user:
	user.TOTPEnabled = true
	user.TOTPSecret = enrollment.Secret

	//	Serialize user to JSON format
	encoded, err := json.Marshal(user)
	if err != nil {
		return user, fmt.Errorf("Problem serializing the data: %s", err)
	}

	//	Save user to the database:
	err = store.systemdb.Update(func(txn *badger.Txn) error {
		err := txn.Set(GetKey("User", user.Name), encoded)
		return err
	})

	//	If there was an error saving the data, report it:
	if err != nil {
		return user, fmt.Errorf("Problem saving the user: %s", err)
	}

	//	Return our updated user:
	return user, nil
}

// GetTOTPEnrollment gets the TOTP enrollment for a user.  If the enrollment information
// can't be found, this will return an error
func (store Manager) GetTOTPEnrollment(userName string) (TotpEnrollment, error) {
	//	Our return item
	enrollment := TotpEnrollment{}

	//	First, make sure we can look up the user's enrollment:
	err := store.systemdb.View(func(txn *badger.Txn) error {
		item, err := txn.Get(GetKey("TotpEnrollment", userName))
		if err != nil {
			return err
		}
		val, err := item.Value()
		if err != nil {
			return err
		}

		if len(val) > 0 {
			//	Unmarshal data into our item
			if err := json.Unmarshal(val, &enrollment); err != nil {
				return err
			}
		}

		return nil
	})

	//	If we got an error, we have a problem:
	if err != nil {
		return enrollment, fmt.Errorf("Enrollment not found")
	}

	//	Return our data:
	return enrollment, nil
}

// GetImage gets the image for an enrollment.
func (enrollment TotpEnrollment) GetImage() ([]byte, error) {
	//	Our return item
	data := []byte{}

	//	Get the key from the url:
	k, err := otp.NewKeyFromURL(enrollment.URL)
	if err != nil {
		return data, fmt.Errorf("There was a problem decoding the enrollment data: %s", err)
	}

	//	Generate an enrollment image
	img, err := k.Image(200, 200)
	if err != nil {
		return data, fmt.Errorf("There was a problem getting the enrollment image: %s", err)
	}

	//	Convert the image to a PNG
	var buf bytes.Buffer
	png.Encode(&buf, img)
	data = buf.Bytes()

	//	Return our image data:
	return data, nil
}
