package test

const Mismatch = "\ngot:\n%s\n\nexpected:\n%s"
const MismatchVal = "\ngot:\n%v\n\nexpected:\n%v"
const MismatchType = "\ngot:\n%T\n\nexpected:\n%T"

const MismatchIn = mismatchIn + Mismatch
const MismatchInVal = mismatchIn + MismatchVal
const MismatchInType = mismatchIn + MismatchType

const Nil = "%s is nil"

const Error = "Error: %s"
const ErrorIn = "Error in %s: %s"

const NotDetected = "%s was not detected"

const Supported = determinedToBe + "supported"
const NotSupported = determinedToBe + "not supported"

const KnownType = determinedToBe + "of an known type"
const UnknownType = determinedToBe + "of an unknown type"

const determinedToBe = "%s was determined to be "
const mismatchIn = "Mismatch in %s"
