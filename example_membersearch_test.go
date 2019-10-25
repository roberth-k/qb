package qb_test

import (
	"fmt"

	"github.com/tetratom/qb"
)

var membersQueryBase = qb.Select("*").From("members")

func BuildQuery(name string, ordered bool) qb.Query {
	q := membersQueryBase

	if name != "" {
		q = q.Where(qb.And("name = ?", name))
	}

	if ordered {
		q = q.OrderBy("created_at DESC")
	}

	return q
}

func Example_searchMembers() {
	allMembers := BuildQuery("", false)
	fmt.Println(allMembers.String())

	allMembersOrdered := BuildQuery("", true)
	fmt.Println(allMembersOrdered.String())

	allMembersJohnDoe := BuildQuery("John Doe", false)
	fmt.Println(allMembersJohnDoe.String())

	allMembersJohnDoeOrdered := BuildQuery("John Doe", true)
	fmt.Println(allMembersJohnDoeOrdered.String())

	// Output:
	// SELECT * FROM members
	// SELECT * FROM members ORDER BY created_at DESC
	// SELECT * FROM members WHERE name = ?
	// SELECT * FROM members WHERE name = ? ORDER BY created_at DESC
}
