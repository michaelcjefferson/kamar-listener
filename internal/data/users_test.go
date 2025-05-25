package data

import (
	"testing"

	"github.com/michaelcjefferson/kamar-listener/internal/assert"
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

func TestGetAll(t *testing.T) {
	type testUser struct {
		ID       int64
		Username string
	}

	tests := []struct {
		name      string
		wantUsers []testUser
	}{
		{
			name: "Get All",
			wantUsers: []testUser{
				{
					ID:       1,
					Username: "test_user_1",
				},
				{
					ID:       2,
					Username: "geoff-the-test",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := newTestUserDB(t)

			m := UserModel{DB: db}

			users, err := m.GetAll()

			assert.NilError(t, err)

			for ind, user := range users {
				assert.Equal(t, user.ID, tt.wantUsers[ind].ID)
				assert.Equal(t, user.Username, tt.wantUsers[ind].Username)
			}
		})
	}
}
