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
New file name may be specified with `-outfile` (see below).

**Usage**

    checksum [options] data_file [dir_to_check]

If no `dir_to_check` is specified, files under the current directory are checked.

**Options**

`-check new|modified|all`

`new` - Calculate checksums for new files only.

`modified` - Calculate checksums for new files,
and for files modified later than the data file.

`all` - Calculate checksums for all files found.

The default is `-check modified` mode.

`-include` - Regular expression for file path to include (e.g. a subdir: `my/data/path`).
Several `-include` parameters may be specified.

`-exclude` - Regular expression for file path to exclude (e.g. a file name: `my_temp_file`).
Several `-exclude` parameters may be specified.

`-delete` - Delete checksums for files not found under the specified folder.
By default, missing files and their checksums are retained in the output data file.

`-outfile` - Save the file checksums to the specified file.
By default, the input file is being rewritten.

`-n` - "Dry run" - do not save anything.

`-v` - Print `OK` messages for all files checked, even with `-check all` option.

`-q` - Don't print `OK` messages for modified files being checked with `-check modified` (default) option.

`-nostat` - Don't print statistics after the check is finished.

**Examples**

    checksum data.sum ~/data

Find all files under `~/data`, and calculate checksums for the files not found in `data.sum`.
Also calculate and update checksums for files newer than `data.sum`.
Save the updated version of `data.sum` file.

    checksum -check all data.sum ~/data

Calculate checksums for all files under `~/data`, and add/update them in `data.sum`.

    checksum -check all -include some/path -exclude tempfile data.sum ~/data

Check all files, whose path includes `some/path`, such as `~/data/some/path`.
Exclude `tempfile` in any subfolder.

    checksum -check all -n -include ^some/path data.sum ~/data

Verify checksums for all files under ~/data/some/path. Keep `data.sum` unchanged.

    checksum -delete data.sum ~/data

Check files under `~/data`, as in the examples above. Remove data for files
not found under `~/data`. Useful to keep `data.sum` "in sync" with actual data.

**Additional note**

It is recommended to keep checksum data file under source control, such as git.
It is helpful to review the changes made by the tools before commiting.
Undesired changes may be reverted easily. Alternatively, save the file under new name,
and check the diff with the original file.
