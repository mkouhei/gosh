History
-------

0.2.0 (not released)
~~~~~~~~~~~~~~~~~~~~

* Enable to re-declare function.
* refactoring of parser import, type, function declaration with go/scanner, go/token instead of regexp.
* Added parser of typeDecl.
* Fixed some bugs.
* Added to execute go get for goimports in Makefile.
* Applied golint, go vet.

0.1.7 (2014-11-28)
~~~~~~~~~~~~~~~~~~

* Supported struct method and pointer parameters and results of function.
* Supported type of function.
* Appended func parser.
* Fixes allowing blanks of the begening of ImportDecl.
* Fixed Installation syntax of README

0.1.6 (2014-11-23)
~~~~~~~~~~~~~~~~~~

* Supported patterns of ImportDecl supported by `go run',
  for example, `[ . | PackageName ] "importPath"' syntax.
* Supported patterns of PackageClause supported by `go run'.

0.1.5 (2014-11-16)
~~~~~~~~~~~~~~~~~~

* Unsupported Go 1.1.
* Added goVersion(), printing license.
* Appended GPLv3 copying permission statement.
* Appended printFlag argument to runCmd().

0.1.4 (2014-11-15)
~~~~~~~~~~~~~~~~~~

* Fixed not work go run when noexistent package in parser.importPkgs.
* Changed log.Printf instead of log.Fatalf when error case at logger().
* Changed appending message string to returns of runCmd().

0.1.3 (2014-11-13)
~~~~~~~~~~~~~~~~~~

* Fixed runtime error occurs when invalid import statement.
* Fixes issue infinite loop of go get.
* Cleanup all working directories on boot.
* Cleard parser.body when non-declaration statement.

0.1.2 (2014-11-12)
~~~~~~~~~~~~~~~~~~

* Changed to print error of runCmd.
* Suppressed "go install: no install location".
* Fixed lacking newline when writing.

0.1.1 (2014-11-10)
~~~~~~~~~~~~~~~~~~

* Fixed deadlock occurs when typing ``Ctrl+D`` immediately after gosh start.
* Fixed fail override tmp code file.

0.1.0 (2014-11-09)
~~~~~~~~~~~~~~~~~~

* First release
