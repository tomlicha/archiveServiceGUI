//source: http://doc.qt.io/qt-5/qtwidgets-widgets-lineedits-example.html

package main

import (
	"bytes"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"math/rand"
	_ "mysql"
	"os"
	"strconv"
	"time"

	"github.com/therecipe/qt/gui"
	"github.com/therecipe/qt/widgets"

	. "ccsdsmo-malgo-examples/archiveservice/archive/consumer"
	. "ccsdsmo-malgo-examples/archiveservice/archive/service"
	. "ccsdsmo-malgo-examples/archiveservice/data"
	. "ccsdsmo-malgo-examples/archiveservice/errors"
	. "ccsdsmo-malgo/com"
	. "ccsdsmo-malgo/mal"
	malapi "ccsdsmo-malgo/mal/api"
)

// Constants for the providers and consumers
const (
	providerURL = "maltcp://127.0.0.1:12400"
	consumerURL = "maltcp://127.0.0.1:14200"
)

// Database ids
const (
	USERNAME = "archiveService"
	PASSWORD = "1a2B3c4D!@?"
	DATABASE = "archive"
	TABLE    = "Archive"
)

const (
	numberOfRows = 80
)

var (
	typesShortForms = []Integer{MAL_FLOAT_TYPE_SHORT_FORM}
	shortForms      = []Long{MAL_FLOAT_SHORT_FORM}
)

// isDatabaseInitialized attribute is true when the database has been initialized
var isDatabaseInitialized = false
var malContext *Context = nil
var clientContext *malapi.ClientContext = nil

func main() {
	testSetup()
	//initTypes()
	widgets.NewQApplication(len(os.Args), os.Args)
	var (
		InitDB     = widgets.NewQGroupBox2("Init DB", nil)
		pushButton = widgets.NewQPushButton2("test init DB", nil)
		DBresp     = widgets.NewQLineEdit(nil)
	)

	var (
		validatorGroup    = widgets.NewQGroupBox2("Operation", nil)
		validatorLabel    = widgets.NewQLabel2("Type:", nil, 0)
		validatorComboBox = widgets.NewQComboBox(nil)
		ID1               = widgets.NewQLineEdit(nil)
		ID2               = widgets.NewQLineEdit(nil)
		ID3               = widgets.NewQLineEdit(nil)
		ID4               = widgets.NewQLineEdit(nil)
		Value             = widgets.NewQLineEdit(nil)
		ButtonOperation   = widgets.NewQPushButton2("Retrieve", nil)
		Result            = widgets.NewQTextEdit(nil)
		Network           = widgets.NewQLineEdit(nil)
		Provider          = widgets.NewQLineEdit(nil)
		Related           = widgets.NewQLineEdit(nil)
		Source            = widgets.NewQLineEdit(nil)
		startTime         = widgets.NewQLineEdit(nil)
		endTime           = widgets.NewQLineEdit(nil)
		sortOrder         = widgets.NewQCheckBox2("Sort Order", nil)
		sortFieldName     = widgets.NewQCheckBox2("Sort Field Name", nil)
	)
	validatorComboBox.AddItems([]string{"Retrieve", "Store", "Count", "Query"})
	ID1.SetPlaceholderText("identificator 1")
	ID2.SetPlaceholderText("identificator 2")
	ID3.SetPlaceholderText("identificator 3")
	ID4.SetPlaceholderText("identificator 4")
	Related.SetPlaceholderText("relation")
	Source.SetPlaceholderText("source")
	startTime.SetPlaceholderText("start-time")
	endTime.SetPlaceholderText("end-time")
	ID1.SetText("en")
	ID2.SetText("cnes")
	ID3.SetText("archiveservice")
	Network.SetPlaceholderText("network")
	Provider.SetPlaceholderText("provider")
	Value.SetValidator(gui.NewQDoubleValidator(Value))
	Value.SetPlaceholderText("value to store")

	Related.SetReadOnly(true)
	Source.SetReadOnly(true)
	startTime.SetReadOnly(true)
	endTime.SetReadOnly(true)
	sortOrder.SetCheckable(false)
	sortFieldName.SetCheckable(false)
	var (
		QueryGroup    = widgets.NewQGroupBox2("Query", nil)
		QueryLabel    = widgets.NewQLabel2("Type:", nil, 0)
		QueryComboBox = widgets.NewQComboBox(nil)
		QueryLineEdit = widgets.NewQLineEdit(nil)
	)
	QueryComboBox.AddItems([]string{"Left", "Centered", "Right"})
	QueryLineEdit.SetPlaceholderText("Placeholder Text")

	validatorComboBox.ConnectCurrentIndexChanged(func(index int) {
		validatorChanged(ID1, ID2, ID3, ID4, ButtonOperation, Value, Provider, Network, Related, Source, startTime, endTime, sortOrder,
			sortFieldName, index)
	})
	QueryComboBox.ConnectCurrentIndexChanged(func(index int) { QueryChanged(QueryLineEdit, index) })
	pushButton.ConnectClicked(func(checked bool) {
		err := checkAndInitDatabase()
		if err != nil {
			DBresp.SetText(err.Error())
		} else {
			DBresp.SetText("database well initiated")

		}
	})

	var echoLayout = widgets.NewQGridLayout2()
	echoLayout.AddWidget(pushButton, 0, 0, 0)
	echoLayout.AddWidget3(DBresp, 1, 0, 1, 2, 0)
	InitDB.SetLayout(echoLayout)

	var retrieveLayout = widgets.NewQGridLayout2()
	retrieveLayout.AddWidget(validatorLabel, 0, 0, 0)
	retrieveLayout.AddWidget(validatorComboBox, 0, 1, 0)
	retrieveLayout.AddWidget3(ID1, 1, 1, 1, 2, 0)
	retrieveLayout.AddWidget3(ID2, 1, 3, 1, 2, 0)
	retrieveLayout.AddWidget3(ID3, 1, 5, 1, 2, 0)
	retrieveLayout.AddWidget3(ID4, 1, 7, 1, 2, 0)
	retrieveLayout.AddWidget3(Value, 2, 1, 1, 2, 0)
	retrieveLayout.AddWidget3(Provider, 2, 3, 1, 2, 0)
	retrieveLayout.AddWidget3(Network, 2, 5, 1, 2, 0)
	retrieveLayout.AddWidget3(Related, 3, 1, 1, 2, 0)
	retrieveLayout.AddWidget3(Source, 3, 3, 1, 2, 0)
	retrieveLayout.AddWidget3(startTime, 3, 5, 1, 2, 0)
	retrieveLayout.AddWidget3(endTime, 3, 7, 1, 2, 0)
	retrieveLayout.AddWidget3(sortFieldName, 4, 1, 1, 2, 0)
	retrieveLayout.AddWidget3(sortOrder, 4, 3, 1, 2, 0)
	retrieveLayout.AddWidget(ButtonOperation, 5, 1, 0)
	retrieveLayout.AddWidget(Result, 6, 1, 0)
	ButtonOperation.ConnectClicked(func(checked bool) {
		var params []string
		var paramsQuery []string
		if ID1.Text() != "" {
			params = append(params, ID1.Text())
		}
		if ID2.Text() != "" {
			params = append(params, ID2.Text())
		}
		if ID3.Text() != "" {
			params = append(params, ID3.Text())
		}
		if ID4.Text() != "" {
			params = append(params, ID4.Text())
		}
		if Related.Text() != "" {
			paramsQuery = append(paramsQuery, Related.Text())
		}
		if Source.Text() != "" {
			paramsQuery = append(paramsQuery, Source.Text())
		}
		if startTime.Text() != "" {
			paramsQuery = append(paramsQuery, startTime.Text())
		}
		if endTime.Text() != "" {
			paramsQuery = append(paramsQuery, endTime.Text())
		}
		if validatorComboBox.CurrentIndex() == 0 {
			Result.SetText(TestRetrieveOK(params))
		}
		if validatorComboBox.CurrentIndex() == 1 {
			newValue := Value.Text()
			newProvider := Provider.Text()
			newNetwork := Network.Text()
			Result.SetText(TestStoreOK(params, newValue, newProvider, newNetwork))
		}
		if validatorComboBox.CurrentIndex() == 2 {
			Result.SetText(strconv.FormatInt(int64(Count(params)), 10))
		}
		if validatorComboBox.CurrentIndex() == 3 {
			Result.SetText(strconv.FormatInt(int64(Count(params)), 10))
		}

	})
	validatorGroup.SetLayout(retrieveLayout)

	var QueryLayout = widgets.NewQGridLayout2()
	QueryLayout.AddWidget(QueryLabel, 0, 0, 0)
	QueryLayout.AddWidget(QueryComboBox, 0, 1, 0)
	QueryLayout.AddWidget3(QueryLineEdit, 1, 0, 1, 2, 0)
	QueryGroup.SetLayout(QueryLayout)

	var layout = widgets.NewQGridLayout2()
	layout.AddWidget(InitDB, 0, 0, 0)
	layout.AddWidget(validatorGroup, 1, 0, 0)
	layout.AddWidget(QueryGroup, 2, 0, 0)

	var window = widgets.NewQMainWindow(nil, 0)
	window.SetWindowTitle("Archive Service")

	var centralWidget = widgets.NewQWidget(window, 0)
	centralWidget.SetLayout(layout)
	window.SetCentralWidget(centralWidget)

	window.Show()

	widgets.QApplication_Exec()
}

func validatorChanged(ID1 *widgets.QLineEdit, ID2 *widgets.QLineEdit, ID3 *widgets.QLineEdit, ID4 *widgets.QLineEdit, ButtonOperation *widgets.QPushButton,
	Value *widgets.QLineEdit, Provider *widgets.QLineEdit, Network *widgets.QLineEdit,
	Related *widgets.QLineEdit, Source *widgets.QLineEdit, startTime *widgets.QLineEdit, endTime *widgets.QLineEdit,
	sortOrder *widgets.QCheckBox, sortFieldName *widgets.QCheckBox, index int) {
	switch index {
	case 0:
		{
			Related.SetReadOnly(true)
			Source.SetReadOnly(true)
			startTime.SetReadOnly(true)
			endTime.SetReadOnly(true)
			sortOrder.SetCheckable(false)
			sortFieldName.SetCheckable(false)
			Value.SetReadOnly(true)
			Provider.SetReadOnly(true)
			Network.SetReadOnly(true)
			Value.Clear()
			Provider.Clear()
			Network.Clear()
			ButtonOperation.SetText("Retrieve")
		}

	case 1:
		{
			Value.SetReadOnly(false)
			Provider.SetReadOnly(false)
			Network.SetReadOnly(false)
			ButtonOperation.SetText("Store")

		}

	case 2:
		{

			ButtonOperation.SetText("Count")
		}

	case 3:
		{
			ButtonOperation.SetText("Query")
			Related.SetReadOnly(false)
			Source.SetReadOnly(false)
			startTime.SetReadOnly(false)
			endTime.SetReadOnly(false)
			sortOrder.SetCheckable(true)
			sortFieldName.SetCheckable(true)
		}

	}
	if index != 3 {
		Related.SetReadOnly(true)
		Source.SetReadOnly(true)
		startTime.SetReadOnly(true)
		endTime.SetReadOnly(true)
		sortOrder.SetCheckable(false)
		sortFieldName.SetCheckable(false)
	}

	Related.Clear()
	Source.Clear()
	startTime.Clear()
	endTime.Clear()
}

func QueryChanged(QueryLineEdit *widgets.QLineEdit, index int) {
	switch index {
	case 0:
		{

		}

	case 1:
		{

		}

	case 2:
		{

		}
	}
}

// initDatabase is used to init the database
func initDabase() error {
	rand.Seed(time.Now().UnixNano())

	// Open the database
	db, err := sql.Open("mysql", USERNAME+":"+PASSWORD+"@/"+DATABASE+"?parseTime=true")
	if err != nil {
		return err
	}

	// Validate the connection by pinging it
	err = db.Ping()
	if err != nil {
		return err
	}

	// Create the transaction (we have to use this method to use rollback and commit)
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	var count int64
	err = tx.QueryRow("SELECT MAX(id) FROM " + TABLE).Scan(&count)

	// If maxID's Valid parameter is set to false then it means its value is nil
	if true {
		fmt.Println("table erased")
		// Delete all the elements of the table Archive
		_, err = tx.Exec("DELETE FROM " + TABLE)
		if err != nil {
			return err
		}

		// Reset the AUTO_INCREMENT value
		_, err = tx.Exec("ALTER TABLE " + TABLE + " AUTO_INCREMENT=0")
		if err != nil {
			return err
		}

		// Commit changes
		tx.Commit()
		// Close the connection with the database
		db.Close()

		// Variable that defines the ArchiveService
		var archiveService *ArchiveService
		// Create the Archive Service
		archiveService = archiveService.CreateService().(*ArchiveService)

		// Insert elements in the table Archive for future tests
		var elementList = NewFloatList(0)
		var boolean = NewBoolean(false)
		// Variable for the different networks
		var networks = []*Identifier{
			NewIdentifier("tests/network1"),
			NewIdentifier("tests/network2"),
		}
		// Variable for the different providers
		var providers = []*URI{
			NewURI("tests/provider1"),
			NewURI("tests/provider2"),
		}

		var objectType ObjectType
		var identifierList = IdentifierList([]*Identifier{NewIdentifier("fr"), NewIdentifier("cnes"), NewIdentifier("archiveservice"), NewIdentifier("test")})
		var archiveDetailsList = *NewArchiveDetailsList(0)

		// Create elements
		for i := 0; i < numberOfRows/2; i++ {
			// Create the value

			var signe = float64(rand.Int63n(2))
			if signe == 0 {
				elementList.AppendElement(NewFloat(rand.Float32()))
			} else {
				elementList.AppendElement(NewFloat(-rand.Float32()))
			}
			objectType = ObjectType{
				Area:    UShort(2),
				Service: UShort(3),
				Version: UOctet(1),
				Number:  UShort((*elementList)[i].GetTypeShortForm()),
			}
			// Object instance identifier
			var objectInstanceIdentifier = Long(int64(i + 1))
			// Variables for ArchiveDetailsList
			var objectKey = ObjectKey{
				Domain: identifierList,
				InstId: Long(0),
			}
			var objectID = ObjectId{
				Type: &objectType,
				Key:  &objectKey,
			}
			var objectDetails = ObjectDetails{
				Related: NewLong(0),
				Source:  &objectID,
			}
			var network = networks[rand.Int63n(int64(len(networks)))]
			var timestamp = NewFineTime(time.Now())
			var provider = providers[rand.Int63n(int64(len(providers)))]
			archiveDetailsList.AppendElement(NewArchiveDetails(objectInstanceIdentifier, objectDetails, network, timestamp, provider))
		}
		_, errorsList, err := archiveService.Store(providerURL, boolean, objectType, identifierList, archiveDetailsList, elementList)
		if errorsList != nil || err != nil {
			if err != nil {
				return err
			} else if errorsList != nil {
				return errors.New(string(*errorsList.ErrorNumber))
			}
		}

		// Store fourty new elements (total 80 elements)
		identifierList = IdentifierList([]*Identifier{NewIdentifier("en"), NewIdentifier("cnes"), NewIdentifier("archiveservice")})
		for i := 0; i < archiveDetailsList.Size(); i++ {
			var objectInstanceIdentifier = Long(int64(i + 41))
			archiveDetailsList[i].InstId = objectInstanceIdentifier
			archiveDetailsList[i].Details.Source.Key.Domain = identifierList
		}
		_, errorsList, err = archiveService.Store(providerURL, boolean, objectType, identifierList, archiveDetailsList, elementList)
		if errorsList != nil || err != nil {
			if err != nil {
				return err
			} else if errorsList != nil {
				return errors.New(string(*errorsList.ErrorNumber))
			} else {
				return errors.New("UNKNOWN ERROR")
			}
		}
	} else {
		// Commit changes
		tx.Commit()
		// Close the connection with the database
		db.Close()
	}

	return nil
}

// checkAndInitDatabase Checks if the Archive table is initialized or not
// If not, it initializes it and inserts datas in the table Archive
func checkAndInitDatabase() error {
	if !isDatabaseInitialized {
		err := initDabase()
		if err != nil {
			return err
		}
		isDatabaseInitialized = true
	}
	return nil
}

func testSetup() error {
	dfltConsumerURL := consumerURL
	malContext, err := NewContext(dfltConsumerURL)
	if err != nil {
		fmt.Printf("error creating MAL context for URI %s: %s", dfltConsumerURL, err)
		return err
	}
	clientContext, err = malapi.NewClientContext(malContext, "test")
	if err != nil {
		fmt.Printf("error creating client context: %s", err)
		return err
	}
	InitMalContext(clientContext)
	return nil
}

func initTypes() {
	for i := 0; i < len(typesShortForms); i++ {
		var objectType = ObjectType{
			Area:    UShort(2),
			Service: UShort(3),
			Version: UOctet(1),
			Number:  UShort(typesShortForms[i]),
		}
		err := objectType.RegisterMALBodyType(shortForms[i])
		if err != nil {
			fmt.Println("%d, cannot register COM object: %s", typesShortForms[i], err.Error())
		}
	}

}
func countDBElement() int64 {
	// Open the database
	db, err := sql.Open("mysql", USERNAME+":"+PASSWORD+"@/"+DATABASE+"?parseTime=true")
	if err != nil {
		return 0
	}

	// Validate the connection by pinging it
	err = db.Ping()
	if err != nil {
		return 0
	}

	// Create the transaction (we have to use this method to use rollback and commit)
	tx, err := db.Begin()
	if err != nil {
		return 0
	}

	// we count the number of item already inserted into the DB
	var count int64
	err = tx.QueryRow("SELECT MAX(id) FROM " + TABLE).Scan(&count)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	return count
}

func TestRetrieveOK(params []string) string {
	// Check if the Archive table is initialized or not
	err := checkAndInitDatabase()
	if err != nil {
		println(err.Error())
	}

	// Variable that defines the ArchiveService
	var archiveService *ArchiveService
	// Create the Archive Service
	service := archiveService.CreateService()
	archiveService = service.(*ArchiveService)
	// Variable that defines the ArchiveService
	var objectType = ObjectType{
		Area:    UShort(2),
		Service: UShort(3),
		Version: UOctet(1),
		Number:  UShort(MAL_FLOAT_TYPE_SHORT_FORM),
	}

	Identifiers := []*Identifier{}
	for i := 0; i < len(params); i++ {
		Identifiers = append(Identifiers, NewIdentifier(params[i]))
	}
	var identifierList = IdentifierList(Identifiers)
	var longList = LongList([]*Long{NewLong(0)})

	// Variables to retrieve the return of this function
	var archiveDetailsList *ArchiveDetailsList
	var elementList ElementList
	var errorsList *ServiceError
	// Start the consumer
	archiveDetailsList, elementList, errorsList, err = archiveService.Retrieve(providerURL, objectType, identifierList, longList)
	if errorsList != nil || err != nil || archiveDetailsList == nil || elementList == nil {
		println(errorsList)
		return "no data found for these identifiers"
	}
	old := os.Stdout // keep backup of the real stdout
	r, w, _ := os.Pipe()
	os.Stdout = w
	for i := 0; i < elementList.Size(); i++ {
		fmt.Println(*(elementList).GetElementAt(i).(*Float))
	}

	outC := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		outC <- buf.String()
	}()

	// back to normal state
	w.Close()
	os.Stdout = old // restoring the real stdout
	out := <-outC

	// reading our temp stdout
	return out
}

func TestStoreOK(params []string, newValue string, newProvider string, newNetwork string) string {
	// Check if the Archive table is initialized or not
	err := checkAndInitDatabase()
	if err != nil {

	}
	itemsinDB := countDBElement()
	fmt.Println(itemsinDB, " items listed in DB")
	// Variables to retrieve the return of this function
	var errorsList *ServiceError
	// Variable that defines the ArchiveService
	var archiveService *ArchiveService
	// Create the Archive Service
	service := archiveService.CreateService()
	archiveService = service.(*ArchiveService)

	// Start the store consumer
	// Create parameters
	// Object that's going to be stored in the archive
	var elementList = NewFloatList(0)
	value, err := strconv.ParseFloat(newValue, 32)
	elementList.AppendElement(NewFloat(float32(value)))
	var boolean = NewBoolean(true)
	var objectType = ObjectType{
		Area:    UShort(2),
		Service: UShort(3),
		Version: UOctet(1),
		Number:  UShort(MAL_FLOAT_TYPE_SHORT_FORM),
	}

	Identifiers := []*Identifier{}
	for i := 0; i < len(params); i++ {
		Identifiers = append(Identifiers, NewIdentifier(params[i]))
	}
	var identifierList = IdentifierList(Identifiers) // Object instance identifier
	var nextID = itemsinDB + 1
	var objectInstanceIdentifier = *NewLong(nextID)
	// Variables for ArchiveDetailsList
	var objectKey = ObjectKey{
		Domain: identifierList,
		InstId: objectInstanceIdentifier,
	}
	var objectID = ObjectId{
		Type: &objectType,
		Key:  &objectKey,
	}
	var objectDetails = ObjectDetails{
		Related: NewLong(1),
		Source:  &objectID,
	}
	var network = NewIdentifier(newNetwork)
	var timestamp = NewFineTime(time.Now())
	var provider = NewURI(newProvider)
	var archiveDetailsList = ArchiveDetailsList([]*ArchiveDetails{NewArchiveDetails(objectInstanceIdentifier, objectDetails, network, timestamp, provider)})

	// Variable to retrieve the return of this function
	var longList *LongList
	// Start the consumer
	longList, errorsList, err = archiveService.Store(providerURL, boolean, objectType, identifierList, archiveDetailsList, elementList)

	if errorsList != nil || err != nil || longList == nil {
		var buffer bytes.Buffer
		buffer.WriteString("error storing data : ")
		buffer.WriteString(err.Error())
		return buffer.String()
	}
	return "data correctly stored"
}

func Count(params []string) Long {

	// Check if the Archive table is initialized or not
	err := checkAndInitDatabase()
	if err != nil {
		return 0
	}

	// Variables to retrieve the return of this function
	var errorsList *ServiceError
	// Variable that defines the ArchiveService
	var archiveService *ArchiveService
	// Create the Archive Service
	service := archiveService.CreateService()
	archiveService = service.(*ArchiveService)

	var objectType = &ObjectType{
		Area:    UShort(2),
		Service: UShort(3),
		Version: UOctet(1),
		Number:  UShort(MAL_FLOAT_TYPE_SHORT_FORM),
	}
	archiveQueryList := NewArchiveQueryList(0)
	Identifiers := []*Identifier{}
	for i := 0; i < len(params); i++ {
		Identifiers = append(Identifiers, NewIdentifier(params[i]))
	}
	var domain = IdentifierList(Identifiers) // Object instance identifier
	archiveQuery := &ArchiveQuery{
		Domain:    &domain,
		Related:   Long(0),
		SortOrder: NewBoolean(true),
	}
	archiveQueryList.AppendElement(archiveQuery)

	var queryFilterList *CompositeFilterSetList

	// Variable to retrieve the return of this function
	var longList *LongList
	// Start the consumer
	longList, errorsList, err = archiveService.Count(providerURL, objectType, archiveQueryList, queryFilterList)

	if errorsList != nil || err != nil || longList == nil {
		return 0
	}
	return *longList.GetElementAt(0).(*Long)
}

func TestQueryOK() {
	// Check if the Archive table is initialized or not
	err := checkAndInitDatabase()
	if err != nil {
		fmt.Println("failed")
	}

	// Variables to retrieve the return of this function
	var errorsList *ServiceError
	// Variable that defines the ArchiveService
	var archiveService *ArchiveService
	// Create the Archive Service
	service := archiveService.CreateService()
	archiveService = service.(*ArchiveService)

	// Create parameters
	var boolean = NewBoolean(true)
	var objectType = ObjectType{
		Area:    UShort(2),
		Service: UShort(3),
		Version: UOctet(1),
		Number:  UShort(MAL_FLOAT_TYPE_SHORT_FORM),
	}
	archiveQueryList := NewArchiveQueryList(0)
	//var domain = IdentifierList([]*Identifier{NewIdentifier("fr"), NewIdentifier("cnes"), NewIdentifier("archiveservice"), NewIdentifier("test")})
	archiveQuery := &ArchiveQuery{
		Related:       Long(0),
		SortOrder:     NewBoolean(true),
		SortFieldName: NewString("domain"),
	}
	archiveQueryList.AppendElement(archiveQuery)
	var queryFilterList *CompositeFilterSetList

	// Variable to retrieve the responses
	var responses []interface{}

	// Start the consumer
	responses, errorsList, err = archiveService.Query(providerURL, boolean, objectType, *archiveQueryList, queryFilterList)
	// Now, verify the responses
	fmt.Println(responses)
	fmt.Println(*responses[3].(*FloatList).GetElementAt(2).(*ObjectDetails))

	if errorsList != nil || err != nil || responses == nil {
		fmt.Println("failed")
	}
}
