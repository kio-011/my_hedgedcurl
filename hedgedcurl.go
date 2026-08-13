    package main

    import (
        "context"
        "flag"
        "fmt"
        "io"
        "net/http"
        "os"
        "time"
    )

    type Info struct {
        err error
        resp *http.Response
    }

    func get(ctx context.Context, ch chan Info, url string) {
        req, err := http.NewRequestWithContext(ctx, "GET", url, nil)

        if err != nil {
            ch <- Info{err : err, resp : nil}
            return
        }

        res, err := http.DefaultClient.Do(req)
        
        if err != nil {
            ch <- Info{err : err, resp : nil}
            return
        }
        select {
        case ch <- Info{err : nil, resp : res}:
        case <-ctx.Done():
            res.Body.Close()
        }
        
    }

    func main () {
        var timeout int
        flag.IntVar(&timeout, "t", 15, "короткий аргумент таймаут запросов в секунду")
        flag.IntVar(&timeout, "timeout", 15, "длинный аргумент таймаут запросов в секунду")

        flag.Parse()

        urls := flag.Args()
        
        if len(urls) == 0 {
            os.Exit(2)
        }

        ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout) * time.Second)
        defer cancel()

        buffer := make(chan Info, len(urls))

        for _, url := range urls {
            go get(ctx, buffer, url)
        }

        count := 0
        for {
            result := <- buffer

            if ctx.Err() == context.DeadlineExceeded {
                os.Exit(228)
            }

            if result.err != nil {
                count++

                if count == len(urls) {
                    fmt.Fprintln(os.Stderr, "все запросы завершились ошибкой")
                    os.Exit(2)
                }

            } else {
                fmt.Println(result.resp.Status)
                fmt.Println(result.resp.Header)

                byts, err := io.ReadAll(result.resp.Body)
                result.resp.Body.Close()

                if err != nil {
                    fmt.Fprintln(os.Stderr, err)
                    os.Exit(2)
                } else {
                    fmt.Println(string(byts))
                }
                
                for {
                    select {
                    case r := <-buffer:
                        if r.resp != nil {
                            r.resp.Body.Close()
                        }
                    default:
                        return
                    }
                }
            }

        }
        
    }