// Copyright 2020 Arthur Sonzogni. All rights reserved.
// Use of this source code is governed by the MIT license that can be found in
// the LICENSE file.
#include <stddef.h>    // for size_t
#include <array>       // for array
#include <atomic>      // for atomic
#include <chrono>      // for operator""s, chrono_literals
#include <cmath>       // for sin
#include <functional>  // for ref, reference_wrapper, function
#include <memory>      // for allocator, shared_ptr, __shared_ptr_access
#include <string>  // for string, basic_string, char_traits, operator+, to_string
#include <thread>   // for sleep_for, thread
#include <utility>  // for move
#include <vector>   // for vector

#include "color_info_sorted_2d.ipp"  // for ColorInfoSorted2D
#include "ftxui/component/component.hpp"  // for Checkbox, Renderer, Horizontal, Vertical, Input, Menu, Radiobox, ResizableSplitLeft, Tab
#include "ftxui/component/component_base.hpp"  // for ComponentBase, Component
#include "ftxui/component/component_options.hpp"  // for MenuOption, InputOption
#include "ftxui/component/event.hpp"              // for Event, Event::Custom
#include "ftxui/component/screen_interactive.hpp"  // for Component, ScreenInteractive
#include "ftxui/dom/elements.hpp"  // for text, color, operator|, bgcolor, filler, Element, vbox, size, hbox, separator, flex, window, graph, EQUAL, paragraph, WIDTH, hcenter, Elements, bold, vscroll_indicator, HEIGHT, flexbox, hflow, border, frame, flex_grow, gauge, paragraphAlignCenter, paragraphAlignJustify, paragraphAlignLeft, paragraphAlignRight, dim, spinner, LESS_THAN, center, yframe, GREATER_THAN
#include "ftxui/dom/flexbox_config.hpp"  // for FlexboxConfig
#include "ftxui/screen/color.hpp"  // for Color, Color::BlueLight, Color::RedLight, Color::Black, Color::Blue, Color::Cyan, Color::CyanLight, Color::GrayDark, Color::GrayLight, Color::Green, Color::GreenLight, Color::Magenta, Color::MagentaLight, Color::Red, Color::White, Color::Yellow, Color::YellowLight, Color::Default, Color::Palette256, ftxui
#include "ftxui/screen/color_info.hpp"  // for ColorInfo
#include "ftxui/screen/terminal.hpp"    // for Size, Dimensions

#include "tiv/tiv_lib.h"
#include "tiv/image_view.hpp"

using namespace ftxui;

int main() {
  auto screen = ScreenInteractive::Fullscreen();

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

  auto paragraph_right = vbox({
        window(text("Align left:"), paragraphAlignLeft(str)),
        window(text("Align center:"), paragraphAlignCenter(str)),
        window(text("Align right:"), paragraphAlignRight(str)),
        window(text("Align justify:"), paragraphAlignJustify(str)),
        window(text("Side by side"), hbox({
                                        paragraph(str),
                                        separator(),
                                        paragraph(str),
                                    })),
        window(text("Elements with different size:"),
                flexbox({
                    make_box(10, 5),
                    make_box(9, 4),
                    make_box(8, 4),
                    make_box(6, 3),
                    make_box(10, 5),
                    make_box(9, 4),
                    make_box(8, 4),
                    make_box(6, 3),
                    make_box(10, 5),
                    make_box(9, 4),
                    make_box(8, 4),
                    make_box(6, 3),
                })),
    });

  // ---------------------------------------------------------------------------
  // HTOP
  // ---------------------------------------------------------------------------
  int shift = 0;

  auto my_graph = [&shift](int width, int height) {
    std::vector<int> output(width);
    for (int i = 0; i < width; ++i) {
      float v = 0.5f;
      v += 0.1f * sin((i + shift) * 0.1f);
      v += 0.2f * sin((i + shift + 10) * 0.15f);
      v += 0.1f * sin((i + shift) * 0.03f);
      v *= height;
      output[i] = (int)v;
    }
    return output;
  };

    auto cell = [](const std::string& path){ return ftxui::image_view(path); };


    int displayIndex{0};
    auto catDisplay = Renderer([&]{
        return cell("bin/drake_shut.png");
    });

    auto imgDisplay = vbox({
        cell("bin/drake_shut.png"),
    });

    auto htop = Renderer([&] {
        auto frequency = vbox({
            text("Frequency [Mhz]") | hcenter,
            hbox({
                vbox({
                    text("2400 "),
                    filler(),
                    text("1200 "),
                    filler(),
                    text("0 "),
                }),
                graph(std::ref(my_graph)) | flex,
            }) | flex,
        });

    return hbox({
               imgDisplay | size(WIDTH, EQUAL, screen.dimx() / 3),
               separator(),
               paragraph_right | flex,
           }) |
           flex;
  });



  // ---------------------------------------------------------------------------
  // Tabs
  // ---------------------------------------------------------------------------

  int tab_index = 0;
  std::vector<std::string> tab_entries = {
      "htop",
  };
  auto tab_selection =
      Menu(&tab_entries, &tab_index, MenuOption::HorizontalAnimated());
  auto tab_content = Container::Tab(
      {
          htop,
      },
      &tab_index);

  auto exit_button =
      Button("Exit", [&] { screen.Exit(); }, ButtonOption::Animated());

  auto main_container = Container::Vertical({
      Container::Horizontal({
          tab_selection,
          exit_button,
      }),
      tab_content,
  });

  auto main_renderer = Renderer(main_container, [&] {
    return vbox({
        text("FTXUI Demo") | bold | hcenter,
        hbox({
            tab_selection->Render() | flex,
            exit_button->Render(),
        }),
        tab_content->Render() | flex,
    });
  });

  std::atomic<bool> refresh_ui_continue = true;
  std::thread refresh_ui([&] {
    while (refresh_ui_continue) {
      using namespace std::chrono_literals;
      std::this_thread::sleep_for(0.05s);
      // The |shift| variable belong to the main thread. `screen.Post(task)`
      // will execute the update on the thread where |screen| lives (e.g. the
      // main thread). Using `screen.Post(task)` is threadsafe.
      screen.Post([&] { shift++; });
      // After updating the state, request a new frame to be drawn. This is done
      // by simulating a new "custom" event to be handled.
      screen.Post(Event::Custom);
    }
  });

  screen.Loop(main_renderer);
  refresh_ui_continue = false;
  refresh_ui.join();

  return 0;
}