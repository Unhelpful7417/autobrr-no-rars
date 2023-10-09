from fastapi import FastAPI, HTTPException
from fastapi.responses import ORJSONResponse
from pydantic import BaseModel, AnyHttpUrl, NonNegativeInt
import requests
import libtorrent as lt

class RequestSchema(BaseModel): # Create model for the request body
    url: AnyHttpUrl # The URL supplied by the user. AnyHttpUrl ensures that the user submits a valid HTTP URL
    tolerance: NonNegativeInt = 0 # The number of .rar files the user is willing to tolerate. Optional field that defaults to zero
    model_config = { # Example request payload shown in the Swagger UI accessible at the /docs endpoint
        "json_schema_extra": {
            "examples": [
                {
                    "url": "https://releases.ubuntu.com/22.04/ubuntu-22.04.3-desktop-amd64.iso.torrent",
                    "tolerance": 0,
                }
            ]
        }
    }
class HealthcheckResponse(BaseModel): # Create model for the response from the healthcheck endpoint
    status: str
    class Config:
        json_schema_extra = {
            "example": {
                "status": "string"
            }
        }
class IncomingURLRequest: # Contains the data about an incoming request to verify a .torrent file
    def __init__(self, url, tolerance):
        self.url = url
        self.tolerance = tolerance 

app = FastAPI() # Start the server

@app.get("/healthcheck", response_model=HealthcheckResponse)
async def healthcheck():
    return {"status": "okay"}

# ? Maybe add another function later on to validate a torrent from the hash and/or from a torrent file

# Define the function to validate the torrent file found at the URL
@app.post("/validate-url/", response_class=ORJSONResponse)
async def validate_torrent_by_url(body: RequestSchema):
    """
    Check if the torrent at the URL provided contains any .rar files
    """
    download = IncomingURLRequest(body.url, body.tolerance) # Instantiate the incoming request class
    try:
        download.request = requests.get(download.url, timeout=5) # Sends a request to the URL supplied by the user
    except:                                  # Definitely bad practice to catch all exceptions here but when I try a URL that doesn't
        raise HTTPException(status_code=400, # exist, ConnectionError isn't catching it and I'm tired of trying to figure it out
            detail={
                "msg": f"Connection could not be made to the server",
                "url": f"{body.url}"
            }
        )

    # Check for .rar files in the torrent at the URL
    if download.request.status_code == 200: # Use the connection initiated earlier

        # If final URL points to a torrent file (ignoring any URL parameters) then
        # open the file with libtorrent to get torrent metadata
        if download.request.url.split('?')[0].endswith(".torrent"):
            download.infoObject = lt.torrent_info(download.request.content)
        else:
            raise HTTPException(status_code=400,
                detail={
                    "msg": f"URL does not point to a .torrent file",
                    "url": f"{body.url}"
                }
            )

        # Get all the .rar files from the torrent file and put them into a list
        download.rarList = []
        for file in download.infoObject.files():
            if file.path.endswith(".rar"):
                download.rarList.append(file.path)
        
        if download.rarList:                                            # If the list of .rar files is not empty, then check if the count
            if len(download.rarList) > download.tolerance:              # is within the tolerance. If the count exceeds the tolerance, respond
                # Return the list of .rar files in the response body    # with HTTP 418. If the number falls within the tolerance, then respond
                raise HTTPException(status_code=418,                    # HTTP 200. If no .rar files found, then also respond HTTP 200.
                    detail={
                        "msg": f"Torrent contains too many .rar files",
                        "tolerance": download.tolerance,
                        "rars": download.rarList,
                        "url": f"{download.url}"
                    }
                )
            else:
                return ORJSONResponse(
                    {
                        "msg": "Found .rar files, but amount was within tolerance",
                        "tolerance": download.tolerance,
                        "rarFiles": download.rarList,
                        "url": f"{download.url}"
                    }
                )
        else:
            return ORJSONResponse({"msg": "No .rar files were found","url": f"{download.url}"})
    else:
        raise HTTPException(status_code=download.request.status_code, # If the initial request to the URL supplied by the user does
            detail={                                                  # not get an HTTP 200, then respond back with the error code
                "msg": f"Can't download the torrent requested",       # their request responded with and give a generic excuse.
                "url": f"{body.url}"
            }
        )