package main

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	cmc "github.com/zytzjx/anthenacmc/cmcserverinfo"
	company "github.com/zytzjx/anthenacmc/companysetting"
	Util "github.com/zytzjx/anthenacmc/utils"
)

func addOther2Multpart(val map[string]io.Reader) {
	var configInstall cmc.ConfigInstall //map[string]interface{}
	if err := configInstall.LoadFile("serialconfig.json"); err != nil {
		fmt.Println(err)
		return
	}
	companyid, _ := configInstall.Results[0].GetCompanyID()
	productid, _ := configInstall.Results[0].GetProductID()
	solutionid, _ := configInstall.Results[0].GetSolutionID()
	siteid, _ := configInstall.Results[0].GetSiteID()
	val["companyid"] = strings.NewReader(strconv.Itoa(companyid))
	val["productid"] = strings.NewReader(strconv.Itoa(productid))
	val["solutionid"] = strings.NewReader(strconv.Itoa(solutionid))
	val["siteid"] = strings.NewReader(strconv.Itoa(siteid))
	pcname, _ := Util.GetPCName()
	val["pcname"] = strings.NewReader(pcname)
	mac, _, _ := company.GetLocalPCInfo()
	val["macaddress"] = strings.NewReader(strings.Replace(mac, ":", "", -1))

}

func hashFileMd5(filePath string) (string, error) {
	var returnMD5String string
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil

}

// UploadLog upload log to cmc server
func UploadLog(logfile string) {
	// staticfileserver + uploadlog/
	// curl https://mc.futuredial.com/uploadfile --insecure -u fdus:392potrero
	//-F md5=c115cf6c96fb70afbd7aedd5e4dad432 –F filetype=TransactionLog  -F datetime=20201025T154738
	//-F”pcname=USER-PC” -F”macaddress=F04DA2DCFAF5” -F companyid=41 -F siteid=1 -F productid=2
	//-F file=@/home/test/temp/log_20201025T154738.zip;type=application/octet-stream
	tr := &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
		TLSClientConfig:    &tls.Config{InsecureSkipVerify: true}, //--insecure
	}
	client := &http.Client{Transport: tr}
	md5data, _ := hashFileMd5(logfile)
	//prepare the reader instances to encode
	values := map[string]io.Reader{
		"file":     mustOpen(logfile), // lets assume its this file
		"md5":      strings.NewReader(md5data),
		"filetype": strings.NewReader("TransactionLog"),
		"datetime": strings.NewReader(time.Now().Format("20060102T150405")),
		// "productid": strings.NewReader(productid),
	}

	addOther2Multpart(values)
	err := upload(client, "https://mc.futuredial.com/uploadfile", values)
	if err != nil {
		panic(err)
	}
}

// upload upload log file
func upload(client *http.Client, url string, values map[string]io.Reader) (err error) {
	// Prepare a form that you will submit to that URL.
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range values {
		var fw io.Writer
		if x, ok := r.(io.Closer); ok {
			defer x.Close()
		}
		// Add an image file
		if x, ok := r.(*os.File); ok {
			if fw, err = w.CreateFormFile(key, x.Name()); err != nil {
				return
			}
		} else {
			// Add other fields
			if fw, err = w.CreateFormField(key); err != nil {
				return
			}
		}
		if _, err = io.Copy(fw, r); err != nil {
			return err
		}

	}
	// Don't forget to close the multipart writer.
	// If you don't close it, your request will be missing the terminating boundary.
	w.Close()

	// Now that you have a form, you can submit it to your handler.
	req, err := http.NewRequest("POST", url, &b)
	if err != nil {
		return
	}
	req.SetBasicAuth("fdus", "392potrero")
	// Don't forget to set the content type, this will contain the boundary.
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Submit the request
	res, err := client.Do(req)
	if err != nil {
		return
	}

	// Check the response
	if res.StatusCode != http.StatusOK {
		err = fmt.Errorf("bad status: %s", res.Status)
	}
	return
}

func mustOpen(f string) *os.File {
	r, err := os.Open(f)
	if err != nil {
		panic(err)
	}
	return r
}
