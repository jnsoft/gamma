# gamma

```
go install github.com/go-delve/delve/cmd/dlv@latest
```

```
go mod download
```

## Code Review
### Input
* new input?
* changes in handeling the input
* changes in storing
* changes in how the input is used

### AAA
* Authentication - is the actor who they say they are?
* Authorisation - is the actor allowed to do this?
* Auditing - have we made a note of what happend?

### New dependencis
* Trusted?
* Popular?
* Well maintained?
* Do we need it?

[Check with Snyk Advisor](https://snyk.io/advisor)

## Functional programming principles
Helps produce readable, predictable and testable code
* Pure functions (given the same input always returns the same result and no side effects, or simply the function can be substituted by a lookup table like a mathematical function)
* Referential transparency (any bound variable or subexpression can be replaced by it's value without changing the meaning of the program, regardless of the order in which they are evaluated)
* Immutability (once a value is declared, it is unchangeable)
* Functions as first class entities (functions can be passed as arguments, returned from other functions, stored in data structures and assigned to variables)
* Higher order funtions (functions that takes functions as values or returns functions, like map, filter and reduce)

