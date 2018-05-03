package testdata

var GlobMatch = `{
    "data": {
        "pairs": [
            {
                "response": {
                    "status": 200,
                    "body": "glob match",
                    "encodedBody": false,
                    "headers": {}
                },
                "request": {
                    "body": {
						"globMatch": "*<item field=*>*"
                    }
                }
            }
        ],
        "globalActions": {
            "delays": []
        }
    },
    "meta": {
        "schemaVersion": "v3",
        "hoverflyVersion": "v0.10.2",
        "timeExported": "2017-02-23T12:43:48Z"
    }
}`
