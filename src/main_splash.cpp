// Copyright 2020 Arthur Sonzogni. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.
#include <array>      // for array
#include <atomic>     // for atomic
#include <chrono>     // for operator""s, chrono_literals
#include <cmath>      // for sin
#include <functional> // for ref, reference_wrapper, function
#include <memory>     // for allocator, shared_ptr, __shared_ptr_access
#include <stddef.h>   // for size_t
#include <string> // for string, basic_string, char_traits, operator+, to_string
#include <thread> // for sleep_for, thread
#include <utility> // for move
#include <vector>  // for vector
#include <termios.h>

#include "color_info_sorted_2d.ipp"      // for ColorInfoSorted2D
#include "ftxui/component/component.hpp" // for Checkbox, Renderer, Horizontal, Vertical, Input, Menu, Radiobox, ResizableSplitLeft, Tab
#include "ftxui/component/component_base.hpp"    // for ComponentBase, Component
#include "ftxui/component/component_options.hpp" // for MenuOption, InputOption
#include "ftxui/component/event.hpp"             // for Event, Event::Custom
#include "ftxui/component/screen_interactive.hpp" // for Component, ScreenInteractive
#include "ftxui/dom/elements.hpp" // for text, color, operator|, bgcolor, filler, Element, vbox, size, hbox, separator, flex, window, graph, EQUAL, paragraph, WIDTH, hcenter, Elements, bold, vscroll_indicator, HEIGHT, flexbox, hflow, border, frame, flex_grow, gauge, paragraphAlignCenter, paragraphAlignJustify, paragraphAlignLeft, paragraphAlignRight, dim, spinner, LESS_THAN, center, yframe, GREATER_THAN
#include "ftxui/dom/flexbox_config.hpp" // for FlexboxConfig
#include "ftxui/screen/color.hpp" // for Color, Color::BlueLight, Color::RedLight, Color::Black, Color::Blue, Color::Cyan, Color::CyanLight, Color::GrayDark, Color::GrayLight, Color::Green, Color::GreenLight, Color::Magenta, Color::MagentaLight, Color::Red, Color::White, Color::Yellow, Color::YellowLight, Color::Default, Color::Palette256, ftxui
#include "ftxui/screen/color_info.hpp" // for ColorInfo
#include "ftxui/screen/terminal.hpp"   // for Size, Dimensions

#include "tiv/image_view.hpp"
#include "tiv/tiv_lib.h"

#include "metrics/uptime.h"
#include "metrics/kernel_info.h"
#include "metrics/cpu_info.h"
#include "metrics/os_info.h"
#include "metrics/memory_info.h"
#include "metrics/process_info.h"
#include "metrics/disk_usage.h"

using namespace ftxui;

void flush_input_buffer() {
    tcflush(STDIN_FILENO, TCIFLUSH);  // Flush the input buffer
}

int main() {
  // Primary keeps the content after exiting the splash screen
  // We prefer this over the alternative
  auto screen = ScreenInteractive::FitComponent();

  // ---------------------------------------------------------------------------
  // Paragraph
  // ---------------------------------------------------------------------------
  auto make_box = [](size_t dimx, size_t dimy) {
    std::string title = std::to_string(dimx) + "x" + std::to_string(dimy);
    return window(text(title) | hcenter | bold,
                  text("content") | hcenter | dim) |
           size(WIDTH, EQUAL, dimx) | size(HEIGHT, EQUAL, dimy);
  };

  std::string str =
      "Lorem Ipsum is simply dummy text of the printing and typesetting "
      "industry. Lorem Ipsum has been the industry's standard dummy text "
      "ever since the 1500s, when an unknown printer took a galley of type "
      "and scrambled it to make a type specimen book.";

  CPUInfo cpuInfo;
  std::string cpuInfoString = cpuInfo.getModelAndCores();

  Uptime uptime;
  std::string uptimeString = uptime.getFormattedUptime();

  KernelInfo kernelInfo;
  std::string kernelInfoString = kernelInfo.getKernelVersion();

  OSInfo osInfo;
  std::string osInfoString = osInfo.getOSInfo();

  MemoryInfo memoryInfo;
  memoryInfo.fetchMemoryInfo();
  std::string totalMemoryString = memoryInfo.getFormattedTotalMemory();

  std::vector<ProcessInfo> processes = ProcessInfo::fetchProcesses();
  std::string topProcessString = "";
  if(processes.size() > 0) {
    topProcessString = processes[0].getProcessName() 
      + " (" + processes[0].getRSSMemoryUsageFormat() + ")";
  }

  DiskUsage diskUsage;
  std::vector<std::string> partitions = diskUsage.getPartitions();
  std::vector<Element> elements;
  int count = 0;

  // Add partitions dynamically up to max
  const int maxPartitions = 5;
  for (const auto& partition : partitions) {
      if (count >= maxPartitions) break;  // Stop after printing 5 partitions

      // Get the partition usage information
      std::string partitionInfo = diskUsage.printDiskUsage(partition);
      
      // Only add as an element if partitionInfo is not blank
      if (partitionInfo.empty()) continue;

      elements.push_back(hbox({
          color(Color::RedLight, text(partition + ": ")),
          color(Color::Default, text(partitionInfo)),
        }));
      count++;
  }


  auto paragraph_right = vbox({
      window(text("Server Information"), vbox({
        hbox({
          color(Color::RedLight, text("OS: ")),
          color(Color::Default, text(osInfoString)),
        }),
        hbox({
          color(Color::RedLight, text("Kernel: ")),
          color(Color::Default, text(kernelInfoString)),
        }),
        hbox({
          color(Color::RedLight, text("CPU: ")),
          color(Color::Default, text(cpuInfoString)),
        }),
        hbox({
          color(Color::RedLight, text("Uptime: ")),
          color(Color::Default, text(uptimeString)),
        }),
      })) | xflex,
      window(text("Memory Information"), vbox({
        hbox({
          color(Color::RedLight, text("Usage: ")),
          color(Color::Default, text(totalMemoryString)),
        }),
        hbox({
          color(Color::RedLight, text("Top Usage: ")),
          color(Color::Default, text(topProcessString)),
        }),
      })) | xflex,
      window(text("Disk Information"), vbox(std::move(elements))) | xflex,
      window(text("Server Info"), paragraphAlignLeft("Welcome! This is where you can insert information about your server.")) | xflex,
  });

  // Show the image on the left side of the screen
  auto cell = [](const std::string &path) { return ftxui::image_view(path); };
  auto catDisplay = Renderer([&] { return cell("bin/drake_shut.png"); });
  auto imgDisplay = vbox({
    cell("bin/espresso.png") | flex,
    separator(),
    hbox({
      color(Color::RedLight, text("drakeor")),
      color(Color::Default, text("@")),
      color(Color::RedLight, text("norberta")),
    }) | hcenter,
    color(Color::Default, text("Use jmgr for more info")) | hcenter,
  });

  // Render everything
  auto splash_content = Renderer([&] {
    return hbox({
      imgDisplay | size(WIDTH, LESS_THAN, screen.dimx() / 3) 
        | size(HEIGHT, EQUAL, screen.dimy()),
      separator(),
      paragraph_right,
    });
  });

  // Main rendering
  auto main_container = Container::Vertical({splash_content});

  auto main_renderer = Renderer(main_container, [&] {
    return vbox({
      // Buffer
        text(" ") | bold | hcenter,
        splash_content->Render() | flex,
    });
  });

  auto programWaitTime = std::chrono::milliseconds(10);
  auto maxProgramRunTime = std::chrono::milliseconds(3000);

  uptimeString = "1 day, 2 hours, 3 minutes";

  // Create a separate thread to wait for the UI to start up 
  // and process the splash screen
  std::thread([&screen, &programWaitTime]() {

    // The splash screen shows a skeleton of the main screen at first
    std::this_thread::sleep_for(programWaitTime);

    // Force a UI update after the async tasks in the background finish
    screen.Post(Event::Custom);
    //std::this_thread::sleep_for(uiRefreshTime);

    // Wait for the UI update to finish and then exit
    screen.Post([&screen] {
        screen.Exit();
    });
  }).detach();

  // Thread to forcefully terminate the program after runtime*3 seconds if it hasn't exited
  std::thread([&maxProgramRunTime]() {
    std::this_thread::sleep_for(maxProgramRunTime);
    std::cerr << "Jettison UI seems stuck: forcing termination" << std::endl;
    std::terminate();  // Forcefully terminate the program
  }).detach();

  // Main screen loop
  screen.Loop(main_renderer);

  // Clear any pending input
  // This clears out most of the garbage and prevents garbage 
  // from appearing in the terminal after
  auto uiFlushTime = std::chrono::milliseconds(100);
  std::this_thread::sleep_for(uiFlushTime);
  flush_input_buffer();  


  return 0;
}