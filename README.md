<a id="readme-top"></a>
<!--suppress HtmlDeprecatedAttribute -->
<h1 align="center">
  <b>Pipsi Utilities</b>
</h1>

<p align="center">
  <a href="https://www.codefactor.io/repository/github/SpoilerRules/pipsi-utils">
    <img src="https://www.codefactor.io/repository/github/SpoilerRules/pipsi-utils/badge" alt="CodeFactor">
  </a>
  <a href="https://github.com/SpoilerRules/pipsi-utils/releases">
    <img src="https://img.shields.io/github/downloads/SpoilerRules/pipsi-utils/total" alt="GitHub Downloads">
  </a>
  <a href="LICENSE">
    <img src="https://img.shields.io/badge/license-GPL--3.0-blue.svg" alt="GPL-3.0 License">
  </a>
</p>

<!--suppress HtmlDeprecatedAttribute -->
<p align="center">
  Pipsi Utils is a command-line app that automates Pipsi installations and updates with an interactive interface.
</p>

<p align="center">
  <img src="https://i.imgur.com/6XMGQqN.gif" alt="Pipsi Utilities Showcase" style="max-width: 100%; height: auto;">
</p>

<details>
  <summary>Table of Contents</summary>
  <ul>
    <li><a href="#getting-started">Getting Started</a></li>
    <li><a href="#building-from-source">Building from Source</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#build-instructions">Build Instructions</a></li>
        <li><a href="#notes">Notes</a></li>
      </ul>
    </li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
  </ul>
</details>

## Getting Started

1. **Download the Latest Release**  
   Visit the [releases page](https://github.com/SpoilerRules/pipsi-utils/releases/latest) and download `pipsi-utils_Windows_x86_64.zip`.

2. **Extract the Files**  
   After downloading, extract the zip file. Inside, you'll find `pipsi-utils.exe`.

3. **Move the Executable (Optional)**  
   For easier access and consistent application data, move `pipsi-utils.exe` to a preferred location (e.g., `C:\Desktop\Favorite Apps\pipsi-utils`).  
   **Note:** The tool stores data such as available Pipsi installations in its directory. Running the executable from different locations may result in missing data or duplicated configurations.

4. **Run the Tool**  
   You can launch `pipsi-utils.exe` using one of these methods:
   - **Via Terminal/Powershell:**  
     Open a terminal or PowerShell window, navigate to the directory, and run:
     ```powershell
     .\pipsi-utils.exe
     ```  
   - **Via Right-Click:**  
     Right-click `pipsi-utils.exe` and select **Open**.
   - **Via Double-Click:**  
     Simply double-click `pipsi-utils.exe` to run it.

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Building from Source

### Prerequisites
- **Go 1.24 (64-bit)** installed on Windows ([download](https://go.dev/dl/))
- **Git** for repository cloning

### Build Instructions

1. **Clone the Repository**
   ```powershell
   git clone https://github.com/SpoilerRules/pipsi-utils.git
   cd pipsi-utils
   ```
2. **Build the Binary**

   Dependencies will be automatically fetched by Go Modules. Run:
   ```powershell
   go build -o pipsi-utils.exe
   ```
   This generates `pipsi-utils.exe` in the project root.

3. **Run the Application**  
   Launch the application to verify the build:
   ```powershell
   .\pipsi-utils.exe
   ```

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Notes
Ensure your Go environment is properly configured (`GOPATH`, `GOROOT`, etc.).

If you encounter issues, check your Go version (`go version`) and ensure it matches the prerequisite (Go 1.24).

<p align="right">(<a href="#readme-top">back to top</a>)</p>

## Contributing

Contributions are what make the open source community such an amazing place to learn, inspire, and create. Any contributions you make are **greatly appreciated**.

If you have a suggestion that would make this better, please fork the repo and create a pull request. You can also simply open an issue with the tag "enhancement".
Don't forget to give the project a star! Thanks again!

1. Fork the Project
2. Create your Feature Branch (`git checkout -b feature/AmazingFeature`)
3. Commit your Changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the Branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

<p align="right">(<a href="#readme-top">back to top</a>)</p>

### Top contributors:

<a href="https://github.com/SpoilerRules/pipsi-utils/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=SpoilerRules/pipsi-utils" alt="contrib.rocks image" />
</a>

## License

Distributed under the [GNU General Public License v3.0](https://www.gnu.org/licenses/gpl-3.0.en.html). See the [LICENSE](LICENSE) file for more information.

<p align="right">(<a href="#readme-top">back to top</a>)</p>