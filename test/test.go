package test

const Mismatch = "\ngot:\n%s\n\nexpected:\n%s\n"
const MismatchVal = "\ngot:\n%v\n\nexpected:\n%v\n"
const MismatchPtr = "\ngot:\n%p\n\nexpected:\n%p\n"
const MismatchType = "\ngot:\n%T\n\nexpected:\n%T\n"

const MismatchIn = mismatchIn + Mismatch
const MismatchInVal = mismatchIn + MismatchVal
const MismatchInType = mismatchIn + MismatchType
const MismatchInPtr = mismatchIn + MismatchPtr

const Nil = "%s is nil\n"

const Error = "Error: %s\n"
const ErrorIn = "Error in %s: %s\n"

const NotDetected = "%s was not detected\n"

const Supported = determinedToBe + "supported\n"
const NotSupported = determinedToBe + "not supported\n"

const KnownType = determinedToBe + "of an known type\n"
const UnknownType = determinedToBe + "of an unknown type\n"

const determinedToBe = "%s was determined to be "
const mismatchIn = "Mismatch in %s"
