package auth

import (
    "testing"
)

func TestHashPasswordAndCheck(t *testing.T) {
    // Test cases
    tests := []struct {
        name     string
        password string
        wantErr  bool
    }{
        {
            name:     "valid password",
            password: "mySecurePassword123!",
            wantErr:  false,
        },
        {
            name:     "empty password",
            password: "",
            wantErr:  false, // bcrypt allows empty passwords, but you might want to change this in your actual implementation
        },
        {
            name:     "long password",
            password: "reallyLongPasswordThatIsStillValid123!@#$%^&*()",
            wantErr:  false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // First, test password hashing
            hashedPassword, err := HashPassword(tt.password)
            if (err != nil) != tt.wantErr {
                t.Errorf("HashPassword() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if !tt.wantErr && hashedPassword == "" {
                t.Error("HashPassword() returned empty hash for valid password")
            }

            // Then, test password verification
            err = CheckPasswordHash(tt.password, hashedPassword)
            if err != nil {
                t.Errorf("CheckPasswordHash() failed to verify valid password: %v", err)
            }

            // Test with wrong password
            err = CheckPasswordHash("wrongPassword", hashedPassword)
            if err == nil {
                t.Error("CheckPasswordHash() verified incorrect password")
            }
        })
    }
}

func TestCheckPasswordHash_Invalid(t *testing.T) {
    // Test invalid hash format
    err := CheckPasswordHash("password", "invalid-hash-format")
    if err == nil {
        t.Error("CheckPasswordHash() should fail with invalid hash format")
    }
}
