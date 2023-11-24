package main

import (
	"context"
	"log"
	pb "mailinglist/proto"
	"time"

	"github.com/alexflint/go-arg"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// logResponse logs the response from a gRPC server.
func logResponse(res *pb.EmailResponse, err error) {
	if err != nil {
		log.Fatalf("  error: %v", err)
	}

	if res.EmailEntry == nil {
		log.Printf("  email not found")
	} else {
		log.Printf("  response: %v", res.EmailEntry)
	}
}

// createEmail creates a new email entry using the gRPC CreateEmail RPC method.
func createEmail(client pb.MailingListServiceClient, addr string) *pb.EmailEntry {
	log.Println("create email")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.CreateEmail(ctx, &pb.CreateEmailRequest{EmailAddr: addr})
	logResponse(res, err)

	return res.EmailEntry
}

// getEmail retrieves an email entry using the gRPC GetEmail RPC method.
func getEmail(client pb.MailingListServiceClient, addr string) *pb.EmailEntry {
	log.Println("get email")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.GetEmail(ctx, &pb.GetEmailRequest{EmailAddr: addr})
	logResponse(res, err)

	return res.EmailEntry
}

// getEmailBatch retrieves a batch of email entries using the gRPC GetEmailBatch RPC method.
func getEmailBatch(client pb.MailingListServiceClient, count int, page int) {
	log.Println("get email batch")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.GetEmailBatch(ctx, &pb.GetEmailBatchRequest{Count: int32(count), Page: int32(page)})
	if err != nil {
		log.Fatalf("  error: %v", err)
	}
	log.Println("response:")
	for i := 0; i < len(res.EmailEntries); i++ {
		log.Printf("  item [%v of %v]: %s", i+1, len(res.EmailEntries), res.EmailEntries[i])
	}
}

// updateEmail updates an email entry using the gRPC UpdateEmail RPC method.
func updateEmail(client pb.MailingListServiceClient, entry pb.EmailEntry) *pb.EmailEntry {
	log.Println("update email")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.UpdateEmail(ctx, &pb.UpdateEmailRequest{EmailEntry: &entry})
	logResponse(res, err)

	return res.EmailEntry
}

// deleteEmail deletes an email entry using the gRPC DeleteEmail RPC method.
func deleteEmail(client pb.MailingListServiceClient, addr string) *pb.EmailEntry {
	log.Println("delete email")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	res, err := client.DeleteEmail(ctx, &pb.DeleteEmailRequest{EmailAddr: addr})
	logResponse(res, err)

	return res.EmailEntry
}

// Command-line arguments
var args struct {
	GrpcAddr string `arg:"env:MAILINGLIST_GRPC_ADDR"`
}

func main() {
	arg.MustParse(&args)

	// Set default gRPC server address if not provided
	if args.GrpcAddr == "" {
		args.GrpcAddr = ":8081"
	}

	// Establish a connection to the gRPC server
	conn, err := grpc.Dial(args.GrpcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	client := pb.NewMailingListServiceClient(conn)

	// Perform various actions using the gRPC client
	// Uncomment and modify the following lines based on your requirements.

	// newEmail := createEmail(client, "9999@999.999")
	// newEmail.ConfirmedAt = 10000
	// updateEmail(client, *newEmail)
	// deleteEmail(client, newEmail.Email)

	// Example of retrieving batches of emails
	getEmailBatch(client, 3, 1)
	getEmailBatch(client, 3, 2)
	getEmailBatch(client, 3, 3)
}


