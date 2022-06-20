/*
 * Copyright (c) 2022, MegaEase
 * All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package global

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func TestReverseMap(t *testing.T) {
	m := map[int]string{
		1: "a",
		2: "b",
		3: "c",
	}
	n := ReverseMap(m)
	assert.Equal(t, 1, n["a"])
	assert.Equal(t, 2, n["b"])
	assert.Equal(t, 3, n["c"])
}

func makeCA(path string, subject *pkix.Name) (*x509.Certificate, *rsa.PrivateKey, error) {
	// creating a CA which will be used to sign all of our certificates using the x509 package from the Go Standard Library
	caCert := &x509.Certificate{
		SerialNumber:          big.NewInt(2019),
		Subject:               *subject,
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10*365, 0, 0),
		IsCA:                  true, // <- indicating this certificate is a CA certificate.
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	// generate a private key for the CA
	caKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, nil, err
	}

	// create the CA certificate
	caBytes, err := x509.CreateCertificate(rand.Reader, caCert, caCert, &caKey.PublicKey, caKey)
	if err != nil {
		return nil, nil, err
	}

	// Create the CA PEM files
	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	if err := ioutil.WriteFile(path+"ca.crt", caPEM.Bytes(), 0644); err != nil {
		return nil, nil, err
	}

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caKey),
	})
	if err := ioutil.WriteFile(path+"ca.key", caPEM.Bytes(), 0644); err != nil {
		return nil, nil, err
	}
	return caCert, caKey, nil
}

func makeCert(path string, caCert *x509.Certificate, caKey *rsa.PrivateKey, subject *pkix.Name, name string) error {
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1658),
		Subject:      *subject,
		//IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:     []string{"localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}
	certBytes, err := x509.CreateCertificate(rand.Reader, cert, caCert, &certKey.PublicKey, caKey)
	if err != nil {
		return err
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err := ioutil.WriteFile(path+name+".crt", certPEM.Bytes(), 0644); err != nil {
		return err
	}

	certKeyPEM := new(bytes.Buffer)
	pem.Encode(certKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certKey),
	})
	if err := ioutil.WriteFile(path+name+".key", certKeyPEM.Bytes(), 0644); err != nil {
		return err
	}
	return nil
}

func TestTLS(t *testing.T) {
	_tls := TLS{}
	conn, e := _tls.Config()
	assert.Nil(t, conn)
	assert.Nil(t, e)

	_tls = TLS{
		Insecure: true,
	}
	conn, e = _tls.Config()
	assert.NotNil(t, conn)
	assert.Nil(t, e)

	path := GetWorkDir() + "/certs/"
	os.MkdirAll(path, 0755)
	defer os.RemoveAll(path)

	_tls = TLS{
		CA:   filepath.Join(path, "./ca.crt"),
		Cert: filepath.Join(path, "./test.crt"),
		Key:  filepath.Join(path, "./test.key"),
	}
	conn, e = _tls.Config()
	assert.Nil(t, conn)
	assert.NotNil(t, e)

	subject := pkix.Name{
		Country:            []string{"Earth"},
		Organization:       []string{"MegaEase"},
		OrganizationalUnit: []string{"Engineering"},
		Locality:           []string{"Mountain"},
		Province:           []string{"Asia"},
		StreetAddress:      []string{"Bridge"},
		PostalCode:         []string{"123456"},
		SerialNumber:       "",
		CommonName:         "CA",
		Names:              []pkix.AttributeTypeAndValue{},
		ExtraNames:         []pkix.AttributeTypeAndValue{},
	}
	caCert, caKey, err := makeCA(path, &subject)
	if err != nil {
		t.Fatalf("make CA Certificate error! - %v", err)
	}
	t.Log("Create the CA certificate successfully.")

	subject.CommonName = "Server"
	subject.Organization = []string{"Server Company"}
	if err := makeCert(path, caCert, caKey, &subject, "test"); err != nil {
		t.Fatal("make Server Certificate error!")
	}
	t.Log("Create and Sign the Server certificate successfully.")

	conn, e = _tls.Config()
	assert.Nil(t, e)
	assert.NotNil(t, conn)

	monkey.Patch(tls.LoadX509KeyPair, func(certFile, keyFile string) (tls.Certificate, error) {
		return tls.Certificate{}, fmt.Errorf("load x509 key pair error")
	})

	conn, e = _tls.Config()
	assert.NotNil(t, e)
	assert.Nil(t, conn)
	monkey.UnpatchAll()
}

func TestNormalize(t *testing.T) {

	// local value
	r := normalize(10, 20, 0, 30)
	assert.Equal(t, 20, r)

	// global value
	r = normalize(10, 0, 0, 10)
	assert.Equal(t, 10, r)

	// default value
	r = normalize(0, 0, 0, 30)
	assert.Equal(t, 30, r)
}

func TestRetry(t *testing.T) {

	r := Retry{
		Times:    3,
		Interval: 100 * time.Millisecond,
	}

	cnt := 0
	f := func() error {
		if cnt < r.Times {
			cnt++
			return fmt.Errorf("error, cnt=%d", cnt)
		}
		return nil
	}

	err := DoRetry("test", "dummy", "tag", r, f)
	assert.NotNil(t, err)
	assert.Equal(t, r.Times, cnt)

	cnt = 1
	err = DoRetry("test", "dummy", "tag", r, f)
	assert.Nil(t, err)
	assert.Equal(t, r.Times, cnt)

}

func TestGetWritableDir(t *testing.T) {
	filename := ""
	dir := MakeDirectory(filename)
	assert.Equal(t, GetWorkDir(), dir)

	filename = "./test.txt"
	dir = MakeDirectory(filename)
	exp, _ := filepath.Abs(filename)
	assert.Equal(t, exp, dir)

	filename = "./none/existed/test.txt"
	exp, _ = filepath.Abs(filename)
	dir = MakeDirectory(filename)
	os.RemoveAll("./none")
	assert.Equal(t, exp, dir)

	filename = "~/none/existed/test.txt"
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.TempDir()
	}
	exp = filepath.Join(home, "none/existed/test.txt")
	dir = MakeDirectory(filename)
	os.RemoveAll(home + "/none")
	assert.Equal(t, exp, dir)
}

func TestGetWorkDirFail(t *testing.T) {
	defer monkey.UnpatchAll()
	monkey.Patch(os.Getwd, func() (string, error) {
		return "", fmt.Errorf("error")
	})

	path := GetWorkDir()
	home, err := os.UserHomeDir()
	assert.Nil(t, err)
	assert.Equal(t, path, home)

	monkey.Patch(os.UserHomeDir, func() (string, error) {
		return "", fmt.Errorf("error")
	})

	path = GetWorkDir()
	assert.Equal(t, path, os.TempDir())

}

func TestMakeDirectoryFail(t *testing.T) {
	defer monkey.UnpatchAll()
	monkey.Patch(os.UserHomeDir, func() (string, error) {
		return "", fmt.Errorf("error")
	})

	filename := "~/test.txt"
	result := MakeDirectory(filename)
	assert.Equal(t, result, filepath.Join(os.TempDir(), filename[2:]))

	monkey.Unpatch(os.UserHomeDir)

	monkey.Patch(filepath.Abs, func(path string) (string, error) {
		return "", fmt.Errorf("error")
	})
	filename = "../test.txt"
	result = MakeDirectory(filename)
	assert.Equal(t, result, filepath.Join(GetWorkDir(), "test.txt"))

	monkey.Unpatch(filepath.Abs)

	monkey.Patch(os.MkdirAll, func(string, os.FileMode) error {
		return fmt.Errorf("error")
	})

	filename = "/not/existed/test.txt"
	result = MakeDirectory(filename)
	assert.Equal(t, result, filepath.Join(GetWorkDir(), "test.txt"))

	monkey.Unpatch(os.MkdirAll)

}

func TestCommandLine(t *testing.T) {
	s := CommandLine("echo", []string{"hello", "world"})
	assert.Equal(t, "echo hello world", s)

	s = CommandLine("kubectl", []string{"get", "pod", "--all-namespaces", "-o", "json"})
	assert.Equal(t, "kubectl get pod --all-namespaces -o json", s)
}
