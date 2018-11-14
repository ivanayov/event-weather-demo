import requests, json, os, sys

def handle(req):

    city = ""
    if len(req) > 0:
        city = req
    else:
        path = os.environ['Http_Path']
        pathArr = path.split("/")
        city = pathArr[1]

    if len(city) == 0:
        sys.exit("Failed to get city from request and/or query")
        return

    with open("/var/openfaas/secrets/weather-api-secret") as f:
        appid = f.read().strip()
        f.close()

    if len(appid) == 0:
        sys.exit("Failed to read appid")
        return

    endpoint = os.getenv("api-endpoint")
    path = os.getenv("api-path")

    query = "?q=" + city + "&appid=" + appid

    url = endpoint+path+query

    res = requests.post(url)
    
    if res.status_code != 200:
        sys.exit("Error accessing wheather api endpoint %s, expected: %d, got: %d\n" % (url, 200, res.status_code))

    weather = res.json()['weather'][0]
    main = res.json()['main']
    wind = res.json()['wind']

    res = {"weather":weather, "main":main, "wind":wind}

    return json.dumps(res)
