package store

import (
	"sort"

	"github.com/cloudmachinery/apps/http-userroles/contracts"
)

type ChangeRoleFunc func(email string, roles Set[string]) error

func adjustRoles(email string, prevRoles, newRoles Set[string], deleteRoleFn, addRoleFn ChangeRoleFunc) error {
	if prevRoles.Equals(newRoles) {
		return nil
	}

	rolesToRemove := prevRoles.Difference(newRoles)
	rolesToAdd := newRoles.Difference(prevRoles)

	if err := deleteRoleFn(email, rolesToRemove); err != nil {
		return err
	}

	if err := addRoleFn(email, rolesToAdd); err != nil {
		return err
	}

	return nil
}

func sortUsers(users []*contracts.User) []*contracts.User {
	sort.Slice(users, func(i, j int) bool {
		return users[i].Email < users[j].Email
	})
	return users
}
