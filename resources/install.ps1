$ErrorActionPreference = "Stop"

# Enable TLS 1.2 since it is required for connections to GitHub.
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

# Helper functions for pretty terminal output.
function Write-Part ([string] $Text) {
  Write-Host $Text -NoNewline
}
function Write-Emphasized ([string] $Text) {
  Write-Host $Text -NoNewLine -ForegroundColor "Yellow"
}
function Write-Done {
  Write-Host " done" -NoNewline -ForegroundColor "Green";
  Write-Host "."
}

# Determine latest release via GitHub API.
$latest_release_uri = "https://api.github.com/repos/netlify/netlify-credential-helper/releases/latest"
Write-Part "Downloading "; Write-Emphasized $latest_release_uri; Write-Part "..."
$latest_release_json = Invoke-WebRequest -Uri $latest_release_uri -UseBasicParsing
Write-Done

Write-Part "Determining latest Netlify Credential release: "
$latest_release = ($latest_release_json | ConvertFrom-Json).tag_name
Write-Emphasized $latest_release; Write-Part "... "
Write-Done

# Create ~\.netlify\helper\bin directory if it doesn't already exist
$install_dir = "${Home}\.netlify\helper\bin"
if (-not (Test-Path $install_dir)) {
  Write-Part "Creating directory "; Write-Emphasized $install_dir; Write-Part "..."
  New-Item -Path $install_dir -ItemType Directory | Out-Null
  Write-Done
}

# Download latest helper release.
$zip_file = "${install_dir}\git-credential-netlify-windows-amd64.zip"
$download_uri = "https://github.com/netlify/netlify-credential-helper/releases/download/" +
                "${latest_release}/git-credential-netlify-windows-amd64.zip"
Write-Part "Downloading "; Write-Emphasized $download_uri; Write-Part "..."
Invoke-WebRequest -Uri $download_uri -OutFile $zip_file -UseBasicParsing
Write-Done

# Extract exe from .zip file.
Write-Part "Extracting "; Write-Emphasized $zip_file
Write-Part " into "; Write-Emphasized ${install_dir}; Write-Part "..."
# Using -Force to overwrite git-credential-netlify if it already exists
Expand-Archive -Path $zip_file -DestinationPath $install_dir -Force
Write-Done

# Remove .zip file.
Write-Part "Removing "; Write-Emphasized $zip_file; Write-Part "..."
Remove-Item -Path $zip_file
Write-Done

# Get Path environment variable for the current user.
$user = [EnvironmentVariableTarget]::User
$path = [Environment]::GetEnvironmentVariable("PATH", $user)

# Check whether the helper is in the Path.
$paths = $path -split ";"
$is_in_path = $paths -contains $install_dir -or $paths -contains "${install_dir}\"

# Add Helper to PATH if it hasn't been added already.
if (-not $is_in_path) {
  Write-Part "Adding "; Write-Emphasized $install_dir; Write-Part " to the "
  Write-Emphasized "PATH"; Write-Part " environment variable..."
  [Environment]::SetEnvironmentVariable("PATH", "${path};${install_dir}", $user)
  # Add Helper to the PATH variable of the current terminal session
  # so `git-credential-netlify` can be used immediately without restarting the 
  # terminal.
  $env:PATH += ";${install_dir}"
  Write-Done
}

Write-Host ""
Write-Host "Netlify Credential Helper for Git was installed successfully." -ForegroundColor "Green"
Write-Host ""
