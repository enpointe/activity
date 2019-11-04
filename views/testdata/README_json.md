# JSON Files

The json files contained within this directory hierachy are used by the tests in the views_test package.
Modifying the contents of these files may result in the breakage of one or more tests.  These files can be recreated by exporting data from the approriate collection via mongoexport

```
$ mongoexport --type json --jsonArray --db <database> --collection <collection> --out user_service_test.json
```

# User Collection Data

**admin_data.json** user entries used by the test suite. This file contains
an entry for the admin user with the following field settings 
entries with the following field settings

| Username | Password     | Privilege |
| -------- | ------------ | --------- |
| admin    | changeMe     | admin     |

**multiuser_data.json** user entries used by the test suite. The current file contains 3 users 
entries with the following field settings

| Username | Password     | Privilege |
| -------- | ------------ | --------- |
| customer1| changeMe     | basic     |
| staff1   | changeMe     | staff     |
| admin1   | changeMe     | admin     |
| customer2| changeMe     | basic     |
| staff2   | changeMe     | staff     |
| admin2   | changeMe     | admin     |





