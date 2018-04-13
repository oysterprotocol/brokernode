package actions

func (as *ActionSuite) Test_IntsJoin_NoInts() {
	v := IntsJoin(nil, " ")

	as.True(v == "")
}

func (as *ActionSuite) Test_IntsJoin_ValidInts() {
	v := IntsJoin([]int{1, 2, 3}, "_")

	as.True(v == "1_2_3")
}

func (as *ActionSuite) Test_IntsJoin_SingleInt() {
	v := IntsJoin([]int{1}, "_")

	as.True(v == "1")
}

func (as *ActionSuite) Test_IntsJoin_InvalidDelim() {
	v := IntsJoin([]int{1, 2, 3}, "")

	as.True(v == "123")
}

func (as *ActionSuite) Test_IntsSplit_InvalidString() {
	v := IntsSplit("abc", " ")

	as.True(v == nil)
}

func (as *ActionSuite) Test_IntsSplit_ValidInput() {
	v := IntsSplit("1_2_3", "_")

	compareIntsArray(as, v, []int{1, 2, 3})
}

func (as *ActionSuite) Test_IntsSplit_SingleInt() {
	v := IntsSplit("1", "_")

	compareIntsArray(as, v, []int{1})
}

func (as *ActionSuite) Test_IntsSplit_MixIntString() {
	v := IntsSplit("1_a_2", "_")

	compareIntsArray(as, v, []int{1, 2})
}

func (as *ActionSuite) Test_IntsSplit_EmptyString() {
	v := IntsSplit("", "_")

	compareIntsArray(as, v, []int{})
}

// Private helper methods
func compareIntsArray(as *ActionSuite, a []int, b []int) {
	as.True(len(a) == len(b))

	for i := 0; i < len(a); i++ {
		as.True(a[i] == b[i])
	}
}
