import traceback
from requests import get
from bs4 import BeautifulSoup


def get_imdb_id_from_keyword(keyword):
    url = 'https://www.imdb.com/find?q={}'.format(keyword)
    response = get(url)
    soup = BeautifulSoup(response.text, 'html.parser')
    result = soup.find('td', class_='result_text')
    if result:
        return result.find('a')['href'].split('/')[2]
    else:
        return None


def get_vidcloud_stream(id, m3u8=False):
    try:
        media_server = (
            BeautifulSoup(
                get(
                    "https://www.2embed.to/embed/imdb/movie?id={}".format(id),
                    headers={"user-agent": "Mozilla/5.0"},
                ).text,
                "html.parser",
            )
            .find("div", class_="media-servers dropdown")
            .find("a")["data-id"]
        )
        recaptcha_resp = get(
            "https://recaptcha.harp.workers.dev/?anchor=https%3A%2F%2Fwww.google.com%2Frecaptcha%2Fapi2%2Fanchor%3Far%3D1%26k%3D6Lf2aYsgAAAAAFvU3-ybajmezOYy87U4fcEpWS4C%26co%3DaHR0cHM6Ly93d3cuMmVtYmVkLnRvOjQ0Mw..%26hl%3Den%26v%3DPRMRaAwB3KlylGQR57Dyk-pF%26size%3Dinvisible%26cb%3D7rsdercrealf&reload=https%3A%2F%2Fwww.google.com%2Frecaptcha%2Fapi2%2Freload%3Fk%3D6Lf2aYsgAAAAAFvU3-ybajmezOYy87U4fcEpWS4C"
        ).json()["rresp"]
        vidcloudresp = get(
            "https://www.2embed.to/ajax/embed/play",
            params={"id": media_server, "_token": recaptcha_resp},
        )
        vid_id = vidcloudresp.json()["link"].split("/")[-1]
        rbstream = "https://rabbitstream.net/embed/m-download/{}".format(
            vid_id
        )
        soup = BeautifulSoup(get(rbstream).text, "html.parser")
        if m3u8:
            recaptcha_2 = get("https://recaptcha.harp.workers.dev/?anchor=https%3A%2F%2Fwww.google.com%2Frecaptcha%2Fapi2%2Fanchor%3Far%3D1%26k%3D6LeLKiYeAAAAAIpuCu6jWvC5X4Y2ZEBd0mlG7h5I%26co%3DaHR0cHM6Ly9yYWJiaXRzdHJlYW0ubmV0OjQ0Mw..%26hl%3Den%26v%3DPRMRaAwB3KlylGQR57Dyk-pF%26size%3Dinvisible%26cb%3D9q8yl3bxqnrf&reload=https%3A%2F%2Fwww.google.com%2Frecaptcha%2Fapi2%2Freload%3Fk%3D6LeLKiYeAAAAAIpuCu6jWvC5X4Y2ZEBd0mlG7h5I").json()["rresp"]
            print(recaptcha_2)
            m3u8_req = get(f"https://rabbitstream.net/ajax/embed-5/getSources?id={vid_id}&_token={recaptcha_2}")
            print(m3u8_req.text)
            m3u8_url = m3u8_req.json()["sources"][0]["file"]
            return m3u8_url
        return [
            a["href"] for a in soup.find("div", class_="download-list").find_all("a")
        ]
    except Exception as e:
        print(e, traceback.format_exc())
        return None


id = get_imdb_id_from_keyword('ABCD 2')
print(get_vidcloud_stream(id, m3u8=True))