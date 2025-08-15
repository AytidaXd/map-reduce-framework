MRAAPS_DIR = src/mrapps
PROJECT_DIR = src/main
PLUGINS = wc.so indexer.so mtiming.so rtiming.so jobcount.so early_exit.so crash.so nocrash.so

all: mrsequential mrcoordinator mrworker $(PLUGINS)

mrsequential:
	(cd $(PROJECT_DIR) && go build $(RACE) mrsequential.go)

mrcoordinator:
	(cd $(PROJECT_DIR) && go build $(RACE) mrcoordinator.go)

mrworker:
	(cd $(PROJECT_DIR) && go build $(RACE) mrworker.go)

%.so:
	(cd $(MRAAPS_DIR) && go build $(RACE) -buildmode=plugin -o $@ $(subst .so,,$@)/main.go)

clean:
	(cd $(MRAAPS_DIR) && go clean)
	(cd $(PROJECT_DIR) && go clean)
	rm -rf src/main/mr-tmp
	rm -f $(MRAAPS_DIR)/*.so
	rm -f $(PROJECT_DIR)/mrsequential $(PROJECT_DIR)/mrcoordinator $(PROJECT_DIR)/mrworker