# simhtape
simhtape is a Go package for handling tape file images in the [standard format](http://simh.trailing-edge.com/docs/simh_magtape.pdf) used by [SimH](http://simh.trailing-edge.com/) and other computer simulators and emulators.

Available functions include...
 * ReadMetaData() and WriteMetaData() for handling headers, trailers, and inter-file gaps
 * ReadRecordData() and WriteRecordData() for handling data blocks (without their associated headers and trailers)
 * Rewind() and SpaceFwd() for positioning the virtual tape image
 * ScanImage() for examining/verifying a tape image file
 * DumpFiles() to extract each file found on the tape image as a numbered blob file 

A sample program is included `simhtapetool` which uses the package to provide a command line interface for...
 * creating a new SimH tape image file (from a CSV definition)
 * dumping each file found on a tape image to a separate blob file
 * scanning the tape image for validity

