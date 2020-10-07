# Remove previously generated .pb.go files
rm -rf /Users/antegulin/git/proto-deployment-go/proto

# Copy all generated proto files to target repo
cp -R ./out/go/bitbucket.org/ag04/proto-deployment-go/proto /Users/antegulin/git/proto-deployment-go

protocOutDir="./out"
targetRepoHome="/Users/antegulin/git/"

function copyFiles {
  for d in */ ; do
    echo "$d"
done
}