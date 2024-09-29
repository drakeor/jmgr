#include <iostream>
#include <ftxui/component/component.hpp>  // for Menu, Renderer, Vertical, Container, Radiobox, Button
#include <ftxui/component/screen_interactive.hpp>  // for ScreenInteractive
#include <ftxui/dom/elements.hpp>  // for text, separator, hbox, vbox, border, flex
#include <string>

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

using namespace ftxui;

int main() {
    auto screen = ScreenInteractive::Fullscreen();


    // ---------------------------------------------------------------------------
    // Overview Tab - Server Selection with RadioBox and Metrics Display
    // ---------------------------------------------------------------------------
    // Server list for the Overview tab using RadioBox for single selection
    std::vector<std::string> server_list = {
        "Server 1 (Online)", "Server 2 (Offline)", "Server 3 (Online)", "Server 4 (Offline)"
    };
    int selected_server = 0;  // First entry selected by default
    auto server_radio_box = Radiobox(&server_list, &selected_server);

    auto serverlist_component = Container::Horizontal({
        server_radio_box,
    });

    auto render_command = [&] {
        Elements line;
        line.push_back(text("Selected Server: ") | bold);
        line.push_back(text(server_list[selected_server]));
        return line;
    };


    auto metrics_panel = Renderer([&] {
        std::string server_name = server_list[selected_server];
        return vbox({
            text("Metrics for " + server_name) | bold | underlined,
            separator(),
            hbox({text("CPU Usage: "), text(server_name == "Server 2 (Offline)" ? "N/A" : "45%")}),
            hbox({text("Memory Usage: "), text(server_name == "Server 2 (Offline)" ? "N/A" : "60%")}),
            hbox({text("Network Traffic: "), text(server_name == "Server 2 (Offline)" ? "N/A" : "120 MB/s")}),
        });
    });

    auto overview_content = Renderer(serverlist_component, [&] {
        auto servers_win = window(text("Servers"),
                                server_radio_box->Render() | vscroll_indicator | frame);
        return hbox({
                vbox({
                    servers_win,
                    filler(),
                }) | size(WIDTH, LESS_THAN, 30),
                window(text("Metrics"), hflow(render_command()) | flex_grow),
            }) |
            flex_grow;
    });

    // Metrics content for the Overview tab (updated dynamically based on selected server)


    // Overview content (left: selectable RadioBox, right: metrics)
    /*auto overview_content = Container::Vertical({
        Renderer([&] {
            return hbox({
                vbox({
                    text("Servers") | bold | underlined,
                    separator(),
                    server_radio_box->Render() | vscroll_indicator | frame | size(WIDTH, EQUAL, 30),  // Scrollable server list using RadioBox
                }) | border | flex,
                separator(),
                metrics_panel->Render() | border | flex,
            });
        })
    });*/

    // ---------------------------------------------------------------------------
    // Servers Tab - Placeholder
    // ---------------------------------------------------------------------------
    auto servers_content = Renderer([] {
        return vbox({
            text("Overview") | bold | underlined,
            text("This is the Servers tab with placeholder content."),
        });
    });

    // ---------------------------------------------------------------------------
    // Tab Content Switching Logic
    // ---------------------------------------------------------------------------
    int tab_index = 0;
    std::vector<std::string> tab_entries = {"Overview", "Servers"};

    auto tab_selection = Menu(&tab_entries, &tab_index, MenuOption::HorizontalAnimated());
    auto tab_content = Container::Tab({
            servers_content,
            overview_content,
        }, &tab_index);

    // ---------------------------------------------------------------------------
    // Exit Button
    // ---------------------------------------------------------------------------
    auto exit_button = Button("Exit", [&] { screen.Exit(); }, ButtonOption::Animated());

    // ---------------------------------------------------------------------------
    // Main Layout and Renderer
    // ---------------------------------------------------------------------------
    auto main_container = Container::Vertical({
        Container::Horizontal({
            tab_selection,
            exit_button,
        }),
        tab_content,
    });

    auto main_renderer = Renderer(main_container, [&] {
        return vbox({
            text("Jettison Server Manager") | bold | hcenter,
            hbox({
                tab_selection->Render() | flex,
                exit_button->Render(),
            }),
            tab_content->Render() | flex,
        });
    });

    // Run the interface
    screen.Loop(main_renderer);

    return 0;
}
