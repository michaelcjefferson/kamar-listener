package data

import (
	"testing"

	"github.com/mjefferson-whs/listener/internal/assert"
)

func TestInsert(t *testing.T) {
	tom := &User{
		Username: "tom",
	}
	err := tom.Password.Set("complicatedpassword")
	assert.NilError(t, err)

	geoff := &User{
		Username: "geoff-the-test",
	}
	err = geoff.Password.Set("anotherpassword")
	assert.NilError(t, err)

	noName := &User{}
	err = noName.Password.Set("password?")
	assert.NilError(t, err)

	noPass := &User{
		Username: "nosh",
	}

	tests := []struct {
		name string
		user *User
		want error
	}{
		{
			name: "Vaild New User",
			user: tom,
			want: nil,
		},
		{
			name: "Duplicate Username",
			user: geoff,
			want: ErrUserAlreadyExists,
		},
		{
			name: "Missing Username",
			user: noName,
			want: ErrMissingUsername,
		},
		{
			name: "Missing Password",
			user: noPass,
			want: ErrMissingPassword,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestUserDB(t)

			m := UserModel{DB: db}

			err := m.Insert(tt.user)

			assert.Equal(t, err, tt.want)
		})
	}
}
