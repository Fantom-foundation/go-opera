package launcher

import (
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/cespare/cp"
)

// These tests are 'smoke tests' for the account related
// subcommands and flags.
//
// For most tests, the test files from package accounts
// are copied into a temporary keystore directory.

func tmpDatadirWithKeystore(t *testing.T) string {
	datadir := tmpdir(t)
	keystore := filepath.Join(datadir, "keystore")
	source := filepath.Join("testdata", "keystore")
	if err := cp.CopyAll(keystore, source); err != nil {
		t.Fatal(err)
	}
	return datadir
}

func TestAccountListEmpty(t *testing.T) {
	cli := exec(t, "--fakenet", "0/1", "account", "list")
	cli.ExpectExit()
}

func TestAccountList(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	cli := exec(t, "--fakenet", "0/1", "account", "list", "--datadir", datadir)

	if runtime.GOOS == "windows" {
		cli.Expect(`
Account #0: {7ef5a6135f1fd6a02593eedc869c6d41d934aef8} keystore://{{.Datadir}}\keystore\UTC--2016-03-22T12-57-55.920751759Z--7ef5a6135f1fd6a02593eedc869c6d41d934aef8
Account #1: {f466859ead1932d743d622cb74fc058882e8648a} keystore://{{.Datadir}}\keystore\aaa
Account #2: {289d485d9771714cce91d3393d764e1311907acc} keystore://{{.Datadir}}\keystore\zzz
`)
	} else {
		cli.Expect(`
Account #0: {7ef5a6135f1fd6a02593eedc869c6d41d934aef8} keystore://{{.Datadir}}/keystore/UTC--2016-03-22T12-57-55.920751759Z--7ef5a6135f1fd6a02593eedc869c6d41d934aef8
Account #1: {f466859ead1932d743d622cb74fc058882e8648a} keystore://{{.Datadir}}/keystore/aaa
Account #2: {289d485d9771714cce91d3393d764e1311907acc} keystore://{{.Datadir}}/keystore/zzz
`)
	}

	cli.ExpectExit()
}

func TestAccountNew(t *testing.T) {
	cli := exec(t, "--fakenet", "0/1", "account", "new", "--lightkdf")

	cli.Expect(`
Your new account is locked with a password. Please give a password. Do not forget this password.
!! Unsupported terminal, password will be echoed.
Passphrase: {{.InputLine "foobar"}}
Repeat passphrase: {{.InputLine "foobar"}}

Your new key was generated
`)
	cli.ExpectRegexp(`
Public address of the key:   0x[0-9a-fA-F]{40}
Path of the secret key file: .*UTC--.+--[0-9a-f]{40}

- You can share your public address with anyone. Others need it to interact with you.
- You must NEVER share the secret key with anyone! The key controls access to your funds!
- You must BACKUP your key file! Without the key, it's impossible to access account funds!
- You must REMEMBER your password! Without the password, it's impossible to decrypt the key!
`)

	cli.ExpectExit()
}

func TestAccountNewBadRepeat(t *testing.T) {
	cli := exec(t, "--fakenet", "0/1", "account", "new", "--lightkdf")

	cli.Expect(`
Your new account is locked with a password. Please give a password. Do not forget this password.
!! Unsupported terminal, password will be echoed.
Passphrase: {{.InputLine "something"}}
Repeat passphrase: {{.InputLine "something else"}}
Fatal: Passphrases do not match
`)
	cli.ExpectExit()
}

func TestAccountUpdate(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	cli := exec(t, "--fakenet", "0/1", "account", "update",
		"--datadir", datadir, "--lightkdf",
		"f466859ead1932d743d622cb74fc058882e8648a")
	defer cli.ExpectExit()
	cli.Expect(`
Unlocking account f466859ead1932d743d622cb74fc058882e8648a | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Passphrase: {{.InputLine "foobar"}}
Please give a new password. Do not forget this password.
Passphrase: {{.InputLine "foobar2"}}
Repeat passphrase: {{.InputLine "foobar2"}}
`)
}

func TestWalletImport(t *testing.T) {
	cli := exec(t, "--fakenet", "0/1", "wallet", "import", "--lightkdf", "testdata/guswallet.json")

	cli.Expect(`
!! Unsupported terminal, password will be echoed.
Passphrase: {{.InputLine "foo"}}
Address: {d4584b5f6229b7be90727b0fc8c6b91bb427821f}
`)

	files, err := ioutil.ReadDir(filepath.Join(cli.Datadir, "keystore"))
	if len(files) != 1 {
		t.Errorf("expected one key file in keystore directory, found %d files (error: %v)", len(files), err)
	}

	cli.ExpectExit()
}

func TestWalletImportBadPassword(t *testing.T) {
	cli := exec(t, "--fakenet", "0/1", "wallet", "import", "--lightkdf", "testdata/guswallet.json")

	cli.Expect(`
!! Unsupported terminal, password will be echoed.
Passphrase: {{.InputLine "wrong"}}
Fatal: could not decrypt key with given password
`)
	cli.ExpectExit()
}

func TestUnlockFlag(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	cli := exec(t,
		"--fakenet", "0/1", "--datadir", datadir, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--unlock", "f466859ead1932d743d622cb74fc058882e8648a",
		"js", "testdata/empty.js")

	cli.Expect(`
Unlocking account f466859ead1932d743d622cb74fc058882e8648a | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Passphrase: {{.InputLine "foobar"}}
`)
	cli.ExpectExit()

	wantMessages := []string{
		"Unlocked account",
		"=0xf466859eAD1932D743d622CB74FC058882E8648A",
	}
	for _, m := range wantMessages {
		if !strings.Contains(cli.StderrText(), m) {
			t.Errorf("stderr text does not contain %q", m)
		}
	}
}

func TestUnlockFlagWrongPassword(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	cli := exec(t,
		"--fakenet", "0/1", "--datadir", datadir, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--unlock", "f466859ead1932d743d622cb74fc058882e8648a")

	cli.Expect(`
Unlocking account f466859ead1932d743d622cb74fc058882e8648a | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Passphrase: {{.InputLine "wrong1"}}
Unlocking account f466859ead1932d743d622cb74fc058882e8648a | Attempt 2/3
Passphrase: {{.InputLine "wrong2"}}
Unlocking account f466859ead1932d743d622cb74fc058882e8648a | Attempt 3/3
Passphrase: {{.InputLine "wrong3"}}
Fatal: Failed to unlock account f466859ead1932d743d622cb74fc058882e8648a (could not decrypt key with given password)
`)
	cli.ExpectExit()
}

// https://github.com/ethereum/go-ethereum/issues/1785
func TestUnlockFlagMultiIndex(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	cli := exec(t,
		"--fakenet", "0/1", "--datadir", datadir, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--unlock", "0,2",
		"js", "testdata/empty.js")

	cli.Expect(`
Unlocking account 0 | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Passphrase: {{.InputLine "foobar"}}
Unlocking account 2 | Attempt 1/3
Passphrase: {{.InputLine "foobar"}}
`)
	cli.ExpectExit()

	wantMessages := []string{
		"Unlocked account",
		"=0x7EF5A6135f1FD6a02593eEdC869c6D41D934aef8",
		"=0x289d485D9771714CCe91D3393D764E1311907ACc",
	}
	for _, m := range wantMessages {
		if !strings.Contains(cli.StderrText(), m) {
			t.Errorf("stderr text does not contain %q", m)
		}
	}
}

func TestUnlockFlagPasswordFile(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	cli := exec(t,
		"--fakenet", "0/1", "--datadir", datadir, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--password", "testdata/passwords.txt", "--unlock", "0,2",
		"js", "testdata/empty.js")

	cli.ExpectExit()

	wantMessages := []string{
		"Unlocked account",
		"=0x7EF5A6135f1FD6a02593eEdC869c6D41D934aef8",
		"=0x289d485D9771714CCe91D3393D764E1311907ACc",
	}
	for _, m := range wantMessages {
		if !strings.Contains(cli.StderrText(), m) {
			t.Errorf("stderr text does not contain %q", m)
		}
	}
}

func TestUnlockFlagPasswordFileWrongPassword(t *testing.T) {
	datadir := tmpDatadirWithKeystore(t)
	cli := exec(t,
		"--fakenet", "0/1", "--datadir", datadir, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--password", "testdata/wrong-passwords.txt", "--unlock", "0,2")

	cli.Expect(`
Fatal: Failed to unlock account 0 (could not decrypt key with given password)
`)
	cli.ExpectExit()
}

func TestUnlockFlagAmbiguous(t *testing.T) {
	store := filepath.Join("testdata", "dupes")
	cli := exec(t,
		"--fakenet", "0/1", "--keystore", store, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--unlock", "f466859ead1932d743d622cb74fc058882e8648a",
		"js", "testdata/empty.js")

	// Helper for the expect template, returns absolute keystore path.
	cli.SetTemplateFunc("keypath", func(file string) string {
		abs, _ := filepath.Abs(filepath.Join(store, file))
		return abs
	})
	cli.Expect(`
Unlocking account f466859ead1932d743d622cb74fc058882e8648a | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Passphrase: {{.InputLine "foobar"}}
Multiple key files exist for address f466859ead1932d743d622cb74fc058882e8648a:
   keystore://{{keypath "1"}}
   keystore://{{keypath "2"}}
Testing your passphrase against all of them...
Your passphrase unlocked keystore://{{keypath "1"}}
In order to avoid this warning, you need to remove the following duplicate key files:
   keystore://{{keypath "2"}}
`)
	cli.ExpectExit()

	wantMessages := []string{
		"Unlocked account",
		"=0xf466859eAD1932D743d622CB74FC058882E8648A",
	}
	for _, m := range wantMessages {
		if !strings.Contains(cli.StderrText(), m) {
			t.Errorf("stderr text does not contain %q", m)
		}
	}
}

func TestUnlockFlagAmbiguousWrongPassword(t *testing.T) {
	store := filepath.Join("testdata", "dupes")
	cli := exec(t,
		"--fakenet", "0/1", "--keystore", store, "--nat", "none", "--nodiscover", "--maxpeers", "0", "--port", "0",
		"--unlock", "f466859ead1932d743d622cb74fc058882e8648a")

	// Helper for the expect template, returns absolute keystore path.
	cli.SetTemplateFunc("keypath", func(file string) string {
		abs, _ := filepath.Abs(filepath.Join(store, file))
		return abs
	})
	cli.Expect(`
Unlocking account f466859ead1932d743d622cb74fc058882e8648a | Attempt 1/3
!! Unsupported terminal, password will be echoed.
Passphrase: {{.InputLine "wrong"}}
Multiple key files exist for address f466859ead1932d743d622cb74fc058882e8648a:
   keystore://{{keypath "1"}}
   keystore://{{keypath "2"}}
Testing your passphrase against all of them...
Fatal: None of the listed files could be unlocked.
`)
	cli.ExpectExit()
}
