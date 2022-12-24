import requests 

if __name__ == "__main__":
    url = 'https://raw.githubusercontent.com/osmosis-labs/assetlists/main/osmosis-1/osmosis-1.assetlist.json'
    resp = requests.get(url)
    
    with open("assetlist.json", "w") as file:
        file.write(resp.text)
        file.close()
