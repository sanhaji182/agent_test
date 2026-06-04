package main

import "github.com/playwright-community/playwright-go"

func main() {
    _ = playwright.LocatorFillOptions{Timeout: playwright.Float(3000)}
    _ = playwright.LocatorClickOptions{Timeout: playwright.Float(3000)}
}
