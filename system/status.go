package system

// TODO: move to Systemctl
// use tabwriter for proper formatting
//
//import (
//	"fmt"
//	"io"
//	"text/tabwriter"
//)
//
//func (s UnitStatus) String() (out string) {
//	out = fmt.Sprintf(
//		`Loaded: %s
//Active: %s`,
//		s.Load, s.Activation)
//
//	b := make([]byte, 1000)
//	if n, _ := u.Read(b); n > 0 {
//		out += fmt.Sprintf("\nLog:\n%s\n", b)
//	}
//
//	return
//}
//
//func (s LoadStatus) String() string {
//	return fmt.Sprintf("%s (%s; %s; %s)",
//		s.Loaded, s.Path, s.State, s.Vendor)
//}
//
//func (s VendorStatus) String() string {
//	return fmt.Sprintf("vendor preset: %s",
//		s.State)
//}
//
//func (s ActivationStatus) String() string {
//	return fmt.Sprintf("%s (%s)",
//		s.State, s.Sub)
//}
//
//func (s SystemStatus) WriteTo(out io.Writer) {
//	tabWriteln(out, s)
//}
//
//func (s UnitStatus) WriteTo(out io.Writer) {
//	tabWriteln(out, s)
//}
//
//func (us Units) WriteTo(out io.Writer) {
//	tabWriteFunc(out, func(w tabwriter.Writer) {
//		fmt.Fprintln(w, "\tunit\t\t\t\tload\tactive\tsub\tdescription")
//
//		for _, u := range us {
//			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t\n",
//				u.Name(), u.Loaded(), u.Active(), u.Sub(), u.Description())
//		}
//	})
//}
//
//func tabWriteFunc(out io.Writer, f func(w tabwriter.Writer)) {
//	w := tabwriter.Writer{}
//
//	w.Init(out, 0, 8, 0, '\t', 0)
//
//	f(w)
//
//	w.Flush()
//}
//func tabWrite(out io.Writer, content Stringer) {
//	tabWriteFunc(out, func(w tabwriter.Writer) {
//		fmt.Fprint(w, content)
//	})
//}
//func tabWriteln(out io.Writer, content Stringer) {
//	tabWriteFunc(out, func(w tabwriter.Writer) {
//		fmt.Fprintln(w, content)
//	})
//}
