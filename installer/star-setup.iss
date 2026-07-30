; Inno Setup Script for Star ⭐ Version Control System
#define MyAppName "Star"
#define MyAppVersion "1.0.0"
#define MyAppPublisher "Star Project"
#define MyAppURL "https://github.com/Luckeris/star"
#define MyAppExeName "star.exe"

[Setup]
AppId={{D37E8863-71AB-4D2A-93B4-9F8C08A1B6F2}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
AppSupportURL={#MyAppURL}
AppUpdatesURL={#MyAppURL}
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
DisableDirPage=no
DisableProgramGroupPage=yes
OutputBaseFilename=star-setup
Compression=lzma
SolidCompression=yes
WizardStyle=modern
WizardImageFile=assets\wizard_banner.bmp
WizardSmallImageFile=assets\wizard_header.bmp
SetupIconFile=assets\star.ico
UninstallDisplayIcon={app}\{#MyAppExeName}
ChangesEnvironment=yes

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Files]
Source: "..\bin\{#MyAppExeName}"; DestDir: "{app}"; Flags: ignoreversion

[Registry]
; Add {app} directory to Current User PATH environment variable
Root: HKCU; Subkey: "Environment"; ValueType: expandsz; ValueName: "Path"; \
    ValueData: "{olddata};{app}"; Check: NeedsAddPath('{app}')

[Code]
// DWM API Constants for Windows 10/11 Immersive Dark Mode
const
  DWMWA_USE_IMMERSIVE_DARK_MODE_BEFORE_20H1 = 19;
  DWMWA_USE_IMMERSIVE_DARK_MODE = 20;

function DwmSetWindowAttribute(
  hwnd: HWND;
  dwAttribute: DWORD;
  var pvAttribute: DWORD;
  cbAttribute: DWORD
): HRESULT;
external 'DwmSetWindowAttribute@dwmapi.dll stdcall delayload';

// Helper to check if Windows system theme is set to Dark Mode
function IsWindowsDarkModeEnabled(): Boolean;
var
  LightMode: DWORD;
begin
  Result := False;
  if RegQueryDWordValue(HKEY_CURRENT_USER, 'Software\Microsoft\Windows\CurrentVersion\Themes\Personalize', 'AppsUseLightTheme', LightMode) then
  begin
    Result := (LightMode = 0);
  end;
end;

procedure ApplyTheme(hWnd: HWND);
var
  DarkMode: DWORD;
begin
  if IsWindowsDarkModeEnabled() then
  begin
    DarkMode := 1;
    if DwmSetWindowAttribute(hWnd, DWMWA_USE_IMMERSIVE_DARK_MODE, DarkMode, SizeOf(DarkMode)) <> 0 then
    begin
      DwmSetWindowAttribute(hWnd, DWMWA_USE_IMMERSIVE_DARK_MODE_BEFORE_20H1, DarkMode, SizeOf(DarkMode));
    end;
  end;
end;

procedure InitializeWizard();
begin
  ApplyTheme(WizardForm.Handle);
end;

// Helper function to check if path is already present in User PATH
function NeedsAddPath(ParamPath: string): boolean;
var
  OrigPath: string;
begin
  if not RegQueryStringValue(HKEY_CURRENT_USER, 'Environment', 'Path', OrigPath) then
  begin
    Result := True;
    exit;
  end;
  Result := Pos(';' + UpperCase(ParamPath) + ';', ';' + UpperCase(OrigPath) + ';') = 0;
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
  ResultCode: Integer;
begin
  if CurStep = ssPostInstall then
  begin
    // Check if Git is installed on user system
    if not Exec('cmd.exe', '/c git --version', '', SW_HIDE, ewWaitUntilTerminated, ResultCode) or (ResultCode <> 0) then
    begin
      MsgBox('Note: Git was not detected on your system PATH.' + #13#10 + #13#10 +
             'Star uses Git under the hood for pushing to GitHub.' + #13#10 +
             'Please install Git from https://git-scm.com/ to use the "star push" feature.',
             mbInformation, MB_OK);
    end;
  end;
end;
