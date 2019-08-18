# Introduction
ModeS Viewer is a simple viewer for the ModeS Go library. It has the ability to connect to either and ADSBSpy device or
replay previously saved data from a text file.

From the data a simple web page is displaying the plane.

# Status
The application is currently in a very rough / early state, but at least it should working.

# Usage
## Run application
### From ADSBSpy

The parameters _adsbSpyServer_ and _adsbSpyPort_ can be used to define the ADSBSpy server. If not defined, default 
values are localhost and port 47806.

```bash
go run cmd/main.go -adsbSpyServer localhost -adsbSpyPort 47806
```

### From File
The file is expected to be a text file, with the same format of data that the data produced by ADSBSpy, for example:

```
...
*8D4070E50053C0000000004F7C40;2464757C;0A;2FC2;
*8D4070E500550000000000C994BE;282A5097;0A;2F95;
*8D4070E5EA3E9866FE1008E87F41;2D4A990B;0A;3E75;
...
```

For replaying a file:
```bash
go run cmd/main.go -adsbSpyServer ./example/example.txt
```

## Connect to the UI
Just open the URL: http://localhost:8081/html/index.html

# Versions
 * v0.1.0: First version

# License

Copyright 2019 Thomas Wuillemin  <thomas.wuillemin@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this project or its content except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.