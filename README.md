# Rounds Proof-of-Concept

Our `create_rounds.go` file has a skeleton implementation of our `CreateRounds` function. This would be called by our BE task runner. FE would then be able to display and interact with these rounds. 

Our `create_rounds_test.go` file has a few test cases, where given some config info and a timestamp, we expect a certain set of rounds to be created. 

This is far from a production implementation, but hopefully it illustrates my thinking. I've left a ton of comments that hopefully walks through the code, step by step. 

More context in the Notion doc.

To run tests, use `go.test`

