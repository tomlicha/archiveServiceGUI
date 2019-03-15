package methods

import (
	"math"
	"math/rand"
	_ "mysql"
	"time"

	_ "mysql"

	. "ccsdsmo-malgo-examples/archiveservice/archive/service"
	. "ccsdsmo-malgo-examples/archiveservice/data"
	. "ccsdsmo-malgo-examples/archiveservice/data/implementation"

	. "ccsdsmo-malgo/com"
	. "ccsdsmo-malgo/mal"
)

var (
	delta float64
)

const (
	providerURL = "maltcp://127.0.0.1:12400"
)

// Store :
func Store() {
	// Variable that defines the ArchiveService
	var archiveService *ArchiveService
	// Create the archive service
	service := archiveService.CreateService()
	archiveService = service.(*ArchiveService)

	step := 50.0
	// Now, we can store the values
	for i := 0; i < 1000; i++ {
		// Increment the delta of the sine
		delta += 1 / step
		// Finally, store this value in the archive
		store(archiveService, float32(math.Sin(float64(delta))), int64(i))

		// Create a random value
		var randValue = time.Duration(rand.Int63n(50) + 1)
		// Wait a little
		time.Sleep(randValue * time.Millisecond / 5)
	}
}

func store(archiveService *ArchiveService, valueOfSine float32, t int64) {
	var elementList = NewSineList(1)
	(*elementList)[0] = NewSine(Long(t), Float(valueOfSine))
	var boolean = NewBoolean(false)
	var objectType = ObjectType{
		Area:    UShort(2),
		Service: UShort(3),
		Version: UOctet(1),
		Number:  UShort(COM_SINE_TYPE_SHORT_FORM),
	}
	var identifierList = IdentifierList([]*Identifier{NewIdentifier("fr"), NewIdentifier("cnes"), NewIdentifier("archiveservice"), NewIdentifier("implementation")})
	// Object instance identifier
	var objectInstanceIdentifier = *NewLong(0)
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
		Related: NewLong(0),
		Source:  &objectID,
	}
	var network = NewIdentifier("network")
	var timestamp = NewFineTime(time.Now())
	var provider = NewURI("main/start")
	var archiveDetailsList = ArchiveDetailsList([]*ArchiveDetails{NewArchiveDetails(objectInstanceIdentifier, objectDetails, network, timestamp, provider)})

	archiveService.Store(providerURL, boolean, objectType, identifierList, archiveDetailsList, elementList)
}
