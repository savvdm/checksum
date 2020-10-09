# checksum
Calculate &amp; update checksum for a set of files.

The program reads the data file, similar to one used by `sha1sum` tool:

    2bd2c36e13d998d111fcc40cb2617f269aa4c01f  Backup/notes/Finance.txt
    df5ba42dcaf7bf7acfe8fde8dc04a7b9508fd3ad  Backup/notes/Golang.txt

It then read all files under the specified directory,
and checks them against the checksums read from the data file.

By default, only files newer than the data file are checked.
This can be controlled by the `-check` option (see below).

New files are always checked, and their checksums are added to the data file.
For existing files, checksums in the data file are updated upon the check.

Modified data file is saved under the same name.
New name may be specified with `-outfile` (see below).

===Usage

    checksum [options] data_file [dir_to_check]

If no `dir_to_check` is specified, files under the current directory are checked.

**Options**

`-check new|modified|all`

`new` - calculate checksums for new files only

`modified` - calculate checksums for new files,
and for files modified later than the data file.

`all` - calculate checksums for all files found.
