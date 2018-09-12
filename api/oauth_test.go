package api

import (
	"testing"
)

func TestAuthHeaderValid_ValidHeader_ReturnsTrue(t *testing.T) {
	//	Arrange
	authHeader := "Bearer SOMELONGTOKEN"

	//	Act
	retval := authHeaderValid(authHeader)

	//	Assert
	if retval == false {
		t.Errorf("AuthHeaderValid indicates bearer token is invalid, but should be valid")
	}
}

func TestAuthHeaderValid_InValidHeader_ReturnsFalse(t *testing.T) {
	//	Arrange
	authHeader := "SOME JANKY STRING"

	//	Act
	retval := authHeaderValid(authHeader)

	//	Assert
	if retval == true {
		t.Errorf("AuthHeaderValid indicates bearer token is valid, but should be invalid")
	}
}

func TestGetTokenFromAuthHeader_ValidBearerToken_ReturnsToken(t *testing.T) {
	//	Arrange
	authHeader := "Bearer YmR1cW82cWQycG0zbTA1dXVoc2c="
	decodedToken := "bduqo6qd2pm3m05uuhsg"

	//	Act
	retval := getTokenFromAuthHeader(authHeader)

	//	Assert
	if retval != decodedToken {
		t.Errorf("getTokenFromAuthHeader should have decoded to %s but got %s instead", decodedToken, retval)
	}
}

func TestBasicHeaderValid_ValidBasicAuth_ReturnsTrue(t *testing.T) {
	//	Arrange
	authHeader := "Basic SOMEENCODEDCREDENTIALS"

	//	Act
	retval := basicHeaderValid(authHeader)

	//	Assert
	if retval == false {
		t.Errorf("basicHeaderValid indicates basic auth is invalid, but should be valid")
	}
}

func TestGetCredentialsFromAuthHeader_ValidBasicAuth_ReturnsCredentials(t *testing.T) {
	//	Arrange
	authHeader := "Basic dGVzdHVzZXI6dGVzdHBhc3N3b3Jk"

	expecteduser := "testuser"
	expectedpassword := "testpassword"

	//	Act
	retuser, retpassword := getCredentialsFromAuthHeader(authHeader)

	//	Assert
	if retuser != expecteduser || retpassword != expectedpassword {
		t.Errorf("getCredentialsFromAuthHeader expected %s / %s but got %s / %s instead", expecteduser, expectedpassword, retuser, retpassword)
	}
}
