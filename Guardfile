def restart_server
  puts "restarting server..."
  system "pkill -f /tmp/go-build"
  system "cd example; go run main.go &"
end

guard :shell do
  watch(%r{\.go$}) {
    restart_server
  }
end
