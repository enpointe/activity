# JSON Files

The json files contained within this directory hierachy are used by the tests in the db_test package.
Modifying the contents of these files may result in the breakage of one or more tests.  These files used
collection fields names for the Collections they are meant to test. These files can be recreated
by exporting data from the approriate collection via mongoexport

```
$ mongoexport --type json --jsonArray --db <database> --collection <collection> --out user_service_test.json
```

# User Collection Data

**user_test.json** user entries used by the test suite. The current file contains 3 users 
entries with the following field settings

| Username | Password     | Privilege |
| -------- | ------------ | --------- |
| customer1| password     | basic     |
| staff    | tellTheTruth | staff     |
| admin    | changeMe     | admin     |

# Exercise Collection Data

**exercise_test.json** exercise entries used by the test suite




