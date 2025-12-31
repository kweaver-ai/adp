#!/usr/bin/env perl
my @result = $_ =~ /\$\(([^\(\)]+)\)/g;
if (!@result) {
    print $_;
} else {
    my $output = $_;
    foreach $original (@result) {
        $envName = uc($original =~ s/\./_/gr);
        if (exists($ENV{$envName})) {
            $envValue = $ENV{$envName};
        } else {
            $envValue = "\${" . $envName . "}";
        }
        $output = $output =~ s/\$\($original\)/$envValue/gr;
    }
    print "$output";
}