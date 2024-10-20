
[![Senior Lab logo](https://cdn.prod.website-files.com/656765cfa435ad4fb32ddb85/656a03754d9a06c0831acade_SeniorLAB_logo-blue.svg)](https://seniorlab.ba)
# SeniorLAB backend

This is the backend part of the code for the **SeniorLAB** application. In it's current form this code can receive a request, and then use **chatGPT-4o-mini,** as well as **google search**, to answer the question the best it can.

## How to use
To deploy this, you need to get the git repo, using the command below:
    
    $ gh repo clone makostrogonac/aibackend
After that, you need to install the requirements for the code to work, using the command below:

    $ pip install -r requirements.txt
 
 After installing everything, to run the program, you use the command:
 

    $ flask run --debug
   We use the --debug flag so that we can be notified of any errors in the code, and to get more info about it and how to fix it.
   This command will run the program on port **5000**. If this port is used by other programs, you can choose a different port using the flag **-p**. 
   For example, if you want to use port **8000**, you use the command
   

    $ flask run --debug -p 8000
If you have any more questions about the code, you can contact me on Discord or Viber for more info.
## Pull requests
All pull requests will be checked by me, so it will not be instant, and it might take a couple of hours for me to check it and to accept it



   
   
